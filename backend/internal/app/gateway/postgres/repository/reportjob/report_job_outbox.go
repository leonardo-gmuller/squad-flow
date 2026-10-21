package reportjobrepo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *ReportJobRepository) ClaimEvents(ctx context.Context, token string, until time.Time) ([]report.QueueEvent, error) {
	rows, err := r.ClaimReportOutbox(ctx, sqlc.ClaimReportOutboxParams{
		PublishToken: token,
		LeaseUntil:   until,
		PageLimit:    25,
	})
	if err != nil {
		return nil, erring.Wrap(err, "reportjob.ClaimEvents")
	}

	events := make([]report.QueueEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, report.QueueEvent{
			ID:          row.ID,
			TraceParent: row.TraceParent,
			TraceState:  row.TraceState,
			JobID:       row.JobID,
			Type:        row.EventType,
		})
	}

	return events, nil
}

func (r *ReportJobRepository) Published(ctx context.Context, id, token string) error {
	affected, err := r.MarkReportEventPublished(ctx, sqlc.MarkReportEventPublishedParams{ID: id, PublishToken: token})
	if err != nil {
		return erring.Wrap(err, "reportjob.Published")
	}

	if affected != 1 {
		return report.ErrLeaseLost
	}

	return nil
}

func (r *ReportJobRepository) Recover(ctx context.Context) error {
	rows, err := r.ListRecoverableReportJobs(ctx)
	if err != nil {
		return erring.Wrap(err, "reportjob.Recover")
	}

	for _, row := range rows {
		err = r.transaction(ctx, func(queries *sqlc.Queries) error {
			recovered, recoverErr := queries.RecoverReportJob(ctx, row.ID)
			if errors.Is(recoverErr, pgx.ErrNoRows) {
				return nil
			}

			if recoverErr != nil {
				return erring.Wrap(recoverErr, "reportjob.Recover.update")
			}

			job, recoverErr := toJob(recovered)
			if recoverErr != nil {
				return recoverErr
			}

			kind := report.EventRequested
			if job.Status == report.JobFailed {
				kind = "report.failed"
			}

			return recordEvent(ctx, queries, job, kind, job.ProgressAgent, job.ProgressMessage, time.Now().UTC())
		})
		if err != nil {
			return err
		}
	}

	return nil
}
