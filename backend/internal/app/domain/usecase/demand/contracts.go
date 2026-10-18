package demand

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
)

type Repository interface {
	Insert(ctx context.Context, item *entity.Demand) error
	GetForUpdate(ctx context.Context, id string) (entity.Demand, error)
	Update(ctx context.Context, item entity.Demand) error
	UpdateDetails(ctx context.Context, item entity.Demand) error
}

type TestRepository interface {
	Insert(ctx context.Context, tc entity.TestCase) error
	CurrentStats(ctx context.Context, id string, cycle int) (passed, failed int, err error)
}

type TransitionRepository interface {
	Insert(ctx context.Context, transition entity.Transition) error
}

type Repositories interface {
	Demands() Repository
	Tests() TestRepository
	Transitions() TransitionRepository
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}

type QueryRepository interface {
	List(ctx context.Context, kind, status, search string) ([]entity.Demand, error)
	Detail(ctx context.Context, id string) (Detail, error)
	AllForMetrics(ctx context.Context) ([]entity.Demand, error)
}

type Detail struct {
	Demand      entity.Demand       `json:"demand"`
	Tests       []entity.TestCase   `json:"tests"`
	Transitions []entity.Transition `json:"transitions"`
}
