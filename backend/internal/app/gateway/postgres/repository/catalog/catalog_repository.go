package catalogrepo

import (
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/catalog"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/uow"
)

type CatalogRepository struct{ *sqlc.Queries }

var _ catalog.Repository = (*CatalogRepository)(nil)

func NewCatalogRepository(db uow.DBTX) *CatalogRepository {
	return &CatalogRepository{Queries: sqlc.New(db)}
}
