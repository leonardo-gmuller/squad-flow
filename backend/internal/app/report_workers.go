package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/config"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	redisqueue "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/queue/redis"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/telemetry"
)

func (a *App) RunReportWorkers(ctx context.Context, cfg config.Config) error {
	if cfg.OpenAI.APIKey == "" {
		slog.Info("report workers disabled: OpenAI key not configured")
		<-ctx.Done()

		return nil
	}

	queue, err := redisqueue.New(cfg.Reports.RedisURL, cfg.Reports.Stream, cfg.Reports.Group, cfg.OpenAI.Timeout+30*time.Second)
	if err != nil {
		return erring.Wrap(err, "app.RunReportWorkers")
	}

	defer func() {
		if closeErr := queue.Close(); closeErr != nil {
			slog.Warn("report queue close failed", "error", closeErr)
		}
	}()

	worker := report.Worker{Repository: a.ReportJobs.Repository, Reports: telemetry.Reports(a.Reports), Queue: telemetry.Events(queue), Timeout: cfg.OpenAI.Timeout, Concurrency: cfg.Reports.Concurrency}
	worker.Run(ctx)

	return nil
}
