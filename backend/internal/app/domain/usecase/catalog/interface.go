package catalog

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
)

// UseCaseInterface defines the operations exposed by the catalog use case.
type UseCaseInterface interface {
	List(ctx context.Context) ([]entity.Workstream, error)
	Members(ctx context.Context) ([]entity.Member, error)
	Create(ctx context.Context, actor entity.User, input Input) (entity.Workstream, error)
	Update(ctx context.Context, actor entity.User, id string, input Input) (entity.Workstream, error)
}

var _ UseCaseInterface = (*UseCase)(nil)
