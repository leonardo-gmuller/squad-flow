package postgres

import (
	"context"
	"embed"
	"fmt"
	"sort"

	"go.opentelemetry.io/otel/metric"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/telemetry"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Client struct {
	Pool    *pgxpool.Pool
	metrics metric.Registration
}

func New(ctx context.Context, url string) (*Client, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, erring.Wrap(err, "postgres.New")
	}

	cfg.ConnConfig.Tracer = telemetry.NewQueryTracer()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, erring.Wrap(err, "postgres.New")
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, erring.Wrap(err, "postgres.New")
	}

	registration, err := telemetry.RegisterPool(pool)
	if err != nil {
		pool.Close()

		return nil, erring.Wrap(err, "postgres.metrics")
	}

	return &Client{Pool: pool, metrics: registration}, nil
}

func (s *Client) Close() {
	if s.metrics != nil {
		_ = s.metrics.Unregister()
	}

	s.Pool.Close()
}

func (s *Client) Migrate(ctx context.Context) error {
	files, err := migrations.ReadDir("migrations")
	if err != nil {
		return erring.Wrap(err, "postgres.Migrate")
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return erring.Wrap(err, "postgres.Migrate")
	}

	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(77621891)`); err != nil {
		return erring.Wrap(err, "postgres.Migrate")
	}

	if _, err = tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		return erring.Wrap(err, "postgres.Migrate")
	}

	for _, file := range files {
		var applied bool

		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, file.Name()).Scan(&applied)
		if err != nil {
			return erring.Wrap(err, "postgres.Migrate")
		}

		if applied {
			continue
		}

		b, err := migrations.ReadFile("migrations/" + file.Name())
		if err != nil {
			return erring.Wrap(err, "postgres.Migrate")
		}

		if _, err = tx.Exec(ctx, string(b), pgx.QueryExecModeSimpleProtocol); err != nil {
			return fmt.Errorf("migration %s: %w", file.Name(), err)
		}

		if _, err = tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, file.Name()); err != nil {
			return erring.Wrap(err, "postgres.Migrate")
		}
	}

	return erring.Wrap(tx.Commit(ctx), "postgres.Migrate")
}
