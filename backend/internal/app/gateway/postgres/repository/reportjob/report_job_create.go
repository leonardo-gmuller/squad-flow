package reportjobrepo

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/telemetry"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

func (r *ReportJobRepository) Enqueue(ctx context.Context, item report.Job) (report.Job, error) {
	var result report.Job

	err := r.transaction(ctx, func(queries *sqlc.Queries) error {
		if err := queries.LockReportUserQueue(ctx, item.UserID); err != nil {
			return erring.Wrap(err, "reportjob.Enqueue.lock")
		}

		existing, err := queries.GetReportJobByRequestKey(ctx, sqlc.GetReportJobByRequestKeyParams{UserID: item.UserID, RequestKey: item.RequestKey})
		if err == nil {
			result, err = toJob(existing)
			if err == nil && result.Request != item.Request {
				return erring.ErrConflict.WithMessage("Idempotency-Key já utilizada para outro pedido")
			}

			return erring.Wrap(err, "reportjob.Enqueue")
		}

		if !errors.Is(err, pgx.ErrNoRows) {
			return erring.Wrap(err, "reportjob.Enqueue.existing")
		}

		payload, err := json.Marshal(item.Request)
		if err != nil {
			return erring.Wrap(err, "reportjob.Enqueue.payload")
		}

		parent, state := telemetry.Inject(ctx)

		row, err := queries.CreateReportJob(ctx, sqlc.CreateReportJobParams{
			ID:          item.ID,
			TraceParent: parent,
			TraceState:  state,
			UserID:      item.UserID,
			RequestKey:  item.RequestKey,
			Request:     payload,
			MaxAttempts: item.MaxAttempts,
		})
		if err != nil {
			var pe *pgconn.PgError
			if errors.As(err, &pe) && pe.Code == "23505" {
				return erring.ErrConflict.WithMessage("você já tem um relatório na fila ou em processamento")
			}

			return erring.Wrap(err, "reportjob.Enqueue")
		}

		result, err = toJob(row)
		if err != nil {
			return erring.Wrap(err, "reportjob.Enqueue")
		}

		return recordEvent(ctx, queries, result, report.EventRequested, "fila", "Solicitação registrada", result.AvailableAt)
	})

	return result, err
}

func recordEvent(ctx context.Context, queries *sqlc.Queries, job report.Job, kind, agent, message string, availableAt time.Time) error {
	id, err := identifier.NewUUID()
	if err != nil {
		return erring.Wrap(err, "reportjob.Enqueue")
	}

	payload, err := json.Marshal(report.JobEvent{
		Agent:   agent,
		Message: message,
		Attempt: job.Attempts,
	})
	if err != nil {
		return erring.Wrap(err, "reportjob.event.payload")
	}

	parent, state := telemetry.Inject(ctx)
	if parent == "" {
		parent, state = job.TraceParent, job.TraceState
	}

	return erring.Wrap(queries.CreateReportEvent(ctx, sqlc.CreateReportEventParams{
		ID:          id,
		JobID:       job.ID,
		TraceParent: parent,
		TraceState:  state,
		EventType:   kind,
		EventData:   payload,
		AvailableAt: availableAt,
	}), "reportjob.event")
}
