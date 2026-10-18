package workpackage

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
)

type Repository interface {
	Insert(ctx context.Context, item entity.WorkPackage) error
	List(ctx context.Context) ([]entity.WorkPackage, error)
	Get(ctx context.Context, id string) (entity.WorkPackage, error)
}
type Repositories interface {
	Packages() Repository
	Demands() demand.Repository
}
type UnitOfWork interface {
	DoPackage(ctx context.Context, callback func(ctx context.Context, repos Repositories) error) error
}
