package packagerepo

import (
	workpackage "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/package"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/uow"
)

type PackageRepository struct{ *sqlc.Queries }

var _ workpackage.Repository = (*PackageRepository)(nil)

func NewPackageRepository(db uow.DBTX) *PackageRepository {
	return &PackageRepository{Queries: sqlc.New(db)}
}
