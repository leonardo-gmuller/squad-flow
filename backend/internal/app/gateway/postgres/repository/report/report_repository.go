package reportrepo

import (
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/uow"
)

type ReportRepository struct{ *sqlc.Queries }

var _ report.Repository = (*ReportRepository)(nil)

func NewReportRepository(db uow.DBTX) *ReportRepository {
	return &ReportRepository{Queries: sqlc.New(db)}
}
