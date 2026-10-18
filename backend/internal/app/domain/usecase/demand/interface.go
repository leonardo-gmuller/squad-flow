package demand

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
)

type UseCaseInterface interface {
	Create(ctx context.Context, input CreateInput) (entity.Demand, error)
	Edit(ctx context.Context, id string, input EditInput) (entity.Demand, error)
	Move(ctx context.Context, id string, input MoveInput) (entity.Demand, error)
	AddTest(ctx context.Context, id string, input AddTestInput) (entity.TestCase, error)
	List(ctx context.Context, kind, status, query string) ([]entity.Demand, error)
	Detail(ctx context.Context, id string) (Detail, error)
	Metrics(ctx context.Context) (Metrics, error)
}

// BoardUseCaseInterface exposes the paginated Kanban board.
//
//nolint:iface // Repository reads and use case validation are distinct architectural boundaries.
type BoardUseCaseInterface interface {
	Board(ctx context.Context, filter BoardFilter) (Board, error)
}

var (
	_ UseCaseInterface      = (*UseCase)(nil)
	_ BoardUseCaseInterface = (*BoardUseCase)(nil)
)
