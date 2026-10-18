package reportrepo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *ReportRepository) Get(ctx context.Context, userID, id string) (report.Report, error) {
	row, err := r.GetReport(ctx, sqlc.GetReportParams{ID: id, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return report.Report{}, erring.ErrNotFound
	}

	if err != nil {
		return report.Report{}, erring.Wrap(err, "reportrepo.Get")
	}

	return toReport(row)
}

func (r *ReportRepository) List(ctx context.Context, userID, demandID string) ([]report.Report, error) {
	rows, err := r.ListReports(ctx, sqlc.ListReportsParams{UserID: userID, DemandID: demandID})
	if err != nil {
		return nil, erring.Wrap(err, "reportrepo.List")
	}

	items := make([]report.Report, 0, len(rows))

	for _, row := range rows {
		item, err := toReport(sqlc.GetReportRow(row))
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}
