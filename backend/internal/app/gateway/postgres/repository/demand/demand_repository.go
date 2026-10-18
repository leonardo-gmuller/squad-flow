package demandrepo

import (
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/uow"
)

type DemandRepository struct{ *sqlc.Queries }

var (
	_ demand.Repository      = (*DemandRepository)(nil)
	_ demand.QueryRepository = (*DemandRepository)(nil)
	_ demand.BoardRepository = (*DemandRepository)(nil)
)

func NewDemandRepository(db uow.DBTX) *DemandRepository {
	return &DemandRepository{Queries: sqlc.New(db)}
}
