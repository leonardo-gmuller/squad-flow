package release

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
)

// UseCaseInterface defines the operations exposed by the release use case.
type UseCaseInterface interface {
	List(ctx context.Context) ([]entity.Release, error)
	Create(ctx context.Context, input Input) (entity.Release, error)
	Update(ctx context.Context, id string, input Input) error
	SetArchived(ctx context.Context, id string, archived bool) error
}

var _ UseCaseInterface = (*UseCase)(nil)
