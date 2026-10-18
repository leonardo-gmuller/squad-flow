package reportjobrepo

import (
	"encoding/json"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func toJob(row sqlc.ReportJob) (report.Job, error) {
	item := report.Job{
		ID:              row.ID,
		UserID:          row.UserID,
		RequestKey:      row.RequestKey,
		Status:          row.Status,
		Attempts:        row.Attempts,
		MaxAttempts:     row.MaxAttempts,
		AvailableAt:     row.AvailableAt,
		ProgressAgent:   row.ProgressAgent,
		ProgressMessage: row.ProgressMessage,
		ErrorMessage:    row.ErrorMessage,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		CompletedAt:     row.CompletedAt,
	}
	if row.ReportID != nil {
		item.ReportID = *row.ReportID
	}

	err := json.Unmarshal(row.Request, &item.Request)

	return item, erring.Wrap(err, "reportjob.toJob")
}

func toEvent(row sqlc.ReportEvent) (report.JobEvent, error) {
	var item report.JobEvent
	err := json.Unmarshal(row.EventData, &item)
	item.ID = row.ID
	item.JobID = row.JobID
	item.Type = row.EventType
	item.CreatedAt = row.CreatedAt

	return item, erring.Wrap(err, "reportjob.toEvent")
}
