package report

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
)

// UseCaseInterface exposes report creation, scope validation and personal history.
type UseCaseInterface interface {
	Validate(ctx context.Context, userID string, input Request) error
	Build(ctx context.Context, userID, reportID string, input Request, progress Progress) (Report, error)
	List(ctx context.Context, userID, demandID string) ([]Report, error)
	Get(ctx context.Context, userID, id string) (Report, error)
}

// JobUseCaseInterface controls asynchronous report requests.
type JobUseCaseInterface interface {
	Submit(ctx context.Context, user entity.User, input Request, requestKey string) (Job, error)
	List(ctx context.Context, userID string) ([]Job, error)
	Get(ctx context.Context, userID, id string) (Job, error)
	Retry(ctx context.Context, user entity.User, id, key string) (Job, error)
}

// WorkerInterface processes and dispatches durable report events.
type WorkerInterface interface {
	Process(ctx context.Context, event QueueEvent) error
	Dispatch(ctx context.Context) error
	Run(ctx context.Context)
}

var (
	_ UseCaseInterface    = (*Service)(nil)
	_ JobUseCaseInterface = (*JobService)(nil)
	_ WorkerInterface     = (*Worker)(nil)
)
