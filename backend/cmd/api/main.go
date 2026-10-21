package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/config"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/telemetry"
)

var version = "dev"

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.New(version)
	if err != nil {
		return erring.Wrap(err, "config")
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer stop()

	telemetryRuntime, err := telemetry.Setup(ctx, cfg)
	if err != nil {
		return erring.Wrap(err, "telemetry setup")
	}

	defer func() {
		flushCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if flushError := telemetryRuntime.Shutdown(flushCtx); flushError != nil {
			slog.Warn("telemetry shutdown failed")
		}
	}()

	// PostgreSQL
	var store *postgres.Client

	for attempt := range 25 {
		store, err = postgres.New(ctx, cfg.Postgres.ConnectionURL())
		if err == nil {
			break
		}

		slog.Warn(
			"waiting for database",
			"attempt", attempt+1,
			"error", err,
		)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(2 * time.Second):
		}
	}

	if err != nil {
		return erring.Wrap(err, "database connection")
	}

	defer store.Close()

	// Migrations
	if migrationErr := store.Migrate(ctx); migrationErr != nil {
		return erring.Wrap(migrationErr, "migration failed")
	}

	// Demo data
	if cfg.Demo.Seed {
		if seedErr := store.Seed(ctx); seedErr != nil {
			return erring.Wrap(seedErr, "demo seed failed")
		}

		if packageSeedErr := store.SeedPackages(ctx); packageSeedErr != nil {
			return erring.Wrap(packageSeedErr, "demo packages seed failed")
		}
	}

	// Application
	application, err := app.New(store)
	if err != nil {
		return erring.Wrap(err, "bootstrap failed")
	}

	if err = application.Configure(ctx, cfg); err != nil {
		return erring.Wrap(err, "security bootstrap failed")
	}

	if cfg.Reports.Role == "worker" {
		return erring.Wrap(application.RunReportWorkers(ctx, cfg), "report workers")
	}

	workersDone := make(chan struct{})

	if cfg.Reports.Role == "all" {
		go func() {
			defer close(workersDone)

			if workerErr := application.RunReportWorkers(ctx, cfg); workerErr != nil {
				slog.Error("report worker stopped", "error", workerErr)
				stop()
			}
		}()
	} else {
		close(workersDone)
	}

	// HTTP API
	httpAPI := api.New(application, cfg)
	defer httpAPI.Close()

	srv := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           httpAPI.Handler,
		ReadTimeout:       cfg.Server.ReadTimeout,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
	}

	go func() {
		slog.Info(
			"squadflow api ready",
			"app", cfg.App.Name,
			"environment", cfg.Environment,
			"version", cfg.App.Version,
			"addr", srv.Addr,
		)

		if err := srv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		cfg.App.GracefulShutdownTimeout,
	)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return erring.Wrap(err, "shutdown")
	}

	select {
	case <-workersDone:
	case <-shutdownCtx.Done():
		slog.Warn("report workers will recover their leases on restart")
	}
	slog.Info("shutdown complete")

	return nil
}
