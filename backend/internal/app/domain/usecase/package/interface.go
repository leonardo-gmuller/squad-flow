package workpackage

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
)

// UseCaseInterface defines the operations exposed by the package use case.
type UseCaseInterface interface {
	Create(ctx context.Context, input Input) (entity.WorkPackage, error)
	List(ctx context.Context) ([]entity.WorkPackage, error)
	Get(ctx context.Context, id string) (entity.WorkPackage, error)
}

var _ UseCaseInterface = (*PackageUseCase)(nil)
