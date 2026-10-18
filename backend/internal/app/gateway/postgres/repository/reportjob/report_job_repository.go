package reportjobrepo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

type ReportJobRepository struct {
	*sqlc.Queries
	pool *pgxpool.Pool
}

var _ report.JobRepository = (*ReportJobRepository)(nil)

func NewReportJobRepository(pool *pgxpool.Pool) *ReportJobRepository {
	return &ReportJobRepository{Queries: sqlc.New(pool), pool: pool}
}

func (r *ReportJobRepository) transaction(ctx context.Context, fn func(*sqlc.Queries) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return erring.Wrap(err, "reportjob.transaction")
	}

	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	if err = fn(r.WithTx(tx)); err != nil {
		return err
	}

	return erring.Wrap(tx.Commit(ctx), "reportjob.commit")
}
