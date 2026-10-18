package reportjobrepo

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *ReportJobRepository) Claim(ctx context.Context, id, token string, until time.Time) (report.Job, error) {
	var result report.Job

	err := r.transaction(ctx, func(queries *sqlc.Queries) error {
		row, err := queries.ClaimReportJob(ctx, sqlc.ClaimReportJobParams{
			ID:         id,
			LeaseToken: token,
			LeaseUntil: until,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return report.ErrLeaseLost
		}

		if err != nil {
			return erring.Wrap(err, "reportjob.Claim")
		}

		result, err = toJob(row)
		if err != nil {
			return err
		}

		return recordEvent(ctx, queries, result, "report.processing", result.ProgressAgent, result.ProgressMessage, time.Now().UTC())
	})

	return result, err
}

func (r *ReportJobRepository) Progress(ctx context.Context, job report.Job, token, agent, message string) error {
	return r.transaction(ctx, func(queries *sqlc.Queries) error {
		affected, err := queries.UpdateReportJobProgress(ctx, sqlc.UpdateReportJobProgressParams{
			ID:         job.ID,
			LeaseToken: token,
			Agent:      agent,
			Message:    message,
		})
		if err != nil {
			return erring.Wrap(err, "reportjob.Progress")
		}

		if affected != 1 {
			return report.ErrLeaseLost
		}

		return recordEvent(ctx, queries, job, "report.progress", agent, message, time.Now().UTC())
	})
}

func (r *ReportJobRepository) Complete(ctx context.Context, job report.Job, token string, result report.Report) error {
	if result.ID != job.ID {
		return erring.ErrInvalidInput
	}

	return r.transaction(ctx, func(queries *sqlc.Queries) error {
		_, err := queries.LockReportJobLease(ctx, sqlc.LockReportJobLeaseParams{ID: job.ID, LeaseToken: token})
		if errors.Is(err, pgx.ErrNoRows) {
			return report.ErrLeaseLost
		}

		if err != nil {
			return erring.Wrap(err, "reportjob.Complete.lock")
		}

		payload, err := json.Marshal(result)
		if err != nil {
			return erring.Wrap(err, "reportjob.Complete.payload")
		}

		if err = queries.CreateReport(ctx, sqlc.CreateReportParams{
			ID:        result.ID,
			UserID:    job.UserID,
			DemandID:  result.DemandID,
			Prompt:    result.Prompt,
			Payload:   payload,
			CreatedAt: result.CreatedAt,
		}); err != nil {
			return erring.Wrap(err, "reportjob.Complete.save")
		}

		affected, err := queries.CompleteReportJob(ctx, sqlc.CompleteReportJobParams{
			ID:         job.ID,
			LeaseToken: token,
			ReportID:   result.ID,
		})
		if err != nil {
			return erring.Wrap(err, "reportjob.Complete")
		}

		if affected != 1 {
			return report.ErrLeaseLost
		}

		return recordEvent(ctx, queries, job, "report.completed", "concluído", "Relatório salvo", time.Now().UTC())
	})
}

func (r *ReportJobRepository) Fail(ctx context.Context, job report.Job, token, message string, next time.Time) error {
	return r.transaction(ctx, func(queries *sqlc.Queries) error {
		row, err := queries.FailReportJob(ctx, sqlc.FailReportJobParams{
			ID:          job.ID,
			LeaseToken:  token,
			Message:     message,
			AvailableAt: next,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return report.ErrLeaseLost
		}

		if err != nil {
			return erring.Wrap(err, "reportjob.Fail")
		}

		failed, err := toJob(row)
		if err != nil {
			return err
		}

		if err = recordEvent(ctx, queries, failed, "report.attempt_failed", "falha", message, time.Now().UTC()); err != nil {
			return err
		}

		if failed.Status == report.JobQueued {
			return recordEvent(ctx, queries, failed, report.EventRequested, "fila", failed.ProgressMessage, next)
		}

		return recordEvent(ctx, queries, failed, "report.failed", "falha", message, time.Now().UTC())
	})
}
