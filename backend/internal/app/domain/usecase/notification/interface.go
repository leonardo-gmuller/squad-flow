package notification

import (
	"context"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
)

// UseCaseInterface defines the operations exposed by the notification use case.
type UseCaseInterface interface {
	List(ctx context.Context, user entity.User, filter string, limit int, before string) (Page, error)
	UnreadCount(ctx context.Context, user entity.User) (int, error)
	Read(ctx context.Context, user entity.User, id string) (time.Time, error)
	ReadAll(ctx context.Context, user entity.User) (time.Time, error)
}

var _ UseCaseInterface = (*UseCase)(nil)
