package reportjobrepo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *ReportJobRepository) Get(ctx context.Context, userID, id string) (report.Job, error) {
	row, err := r.GetReportJob(ctx, sqlc.GetReportJobParams{ID: id, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return report.Job{}, erring.ErrNotFound
	}

	if err != nil {
		return report.Job{}, erring.Wrap(err, "reportjob.Get")
	}

	return toJob(row)
}

func (r *ReportJobRepository) List(ctx context.Context, userID string) ([]report.Job, error) {
	rows, err := r.ListReportJobs(ctx, userID)
	if err != nil {
		return nil, erring.Wrap(err, "reportjob.List")
	}

	items := make([]report.Job, 0, len(rows))

	for _, row := range rows {
		item, err := toJob(row)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *ReportJobRepository) History(ctx context.Context, userID, id string) ([]report.JobEvent, error) {
	rows, err := r.ListReportEvents(ctx, sqlc.ListReportEventsParams{JobID: id, UserID: userID})
	if err != nil {
		return nil, erring.Wrap(err, "reportjob.History")
	}

	events := make([]report.JobEvent, 0, len(rows))

	for _, row := range rows {
		event, err := toEvent(row)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}
