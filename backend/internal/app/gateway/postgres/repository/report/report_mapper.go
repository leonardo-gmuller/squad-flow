package reportrepo

import (
	"encoding/json"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func toReport(row sqlc.GetReportRow) (report.Report, error) {
	var item report.Report
	err := json.Unmarshal(row.Payload, &item)
	item.UserID = row.UserID

	return item, erring.Wrap(err, "reportrepo.toReport")
}
