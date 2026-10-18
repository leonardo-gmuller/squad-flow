package releaserepo

import (
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/release"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/uow"
)

type ReleaseRepository struct{ *sqlc.Queries }

var _ release.Repository = (*ReleaseRepository)(nil)

func NewReleaseRepository(db uow.DBTX) *ReleaseRepository {
	return &ReleaseRepository{Queries: sqlc.New(db)}
}
