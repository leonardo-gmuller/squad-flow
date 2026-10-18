package notificationrepo

import (
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/notification"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/uow"
)

type NotificationRepository struct{ *sqlc.Queries }

var _ notification.Repository = (*NotificationRepository)(nil)

func NewNotificationRepository(db uow.DBTX) *NotificationRepository {
	return &NotificationRepository{Queries: sqlc.New(db)}
}
