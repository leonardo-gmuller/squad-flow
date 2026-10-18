package transitionrepo

import (
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/uow"
)

type TransitionRepository struct{ *sqlc.Queries }

var _ demand.TransitionRepository = (*TransitionRepository)(nil)

func NewTransitionRepository(db uow.DBTX) *TransitionRepository {
	return &TransitionRepository{Queries: sqlc.New(db)}
}
