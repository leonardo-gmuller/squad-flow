package reportrepo

import (
	"context"
	"encoding/json"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *ReportRepository) Save(ctx context.Context, item report.Report) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return erring.Wrap(err, "reportrepo.Save")
	}

	err = r.CreateReport(ctx, sqlc.CreateReportParams{
		ID:        item.ID,
		UserID:    item.UserID,
		DemandID:  item.DemandID,
		Prompt:    item.Prompt,
		Payload:   payload,
		CreatedAt: item.CreatedAt,
	})

	return erring.Wrap(err, "reportrepo.Save")
}
