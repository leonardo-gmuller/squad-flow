package notificationrepo

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/notification"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *NotificationRepository) List(ctx context.Context, userID string, unread bool, limit int, cursor *notification.Cursor) ([]entity.Notification, int, error) {
	params := sqlc.ListNotificationsParams{
		UserID:    userID,
		Unread:    unread,
		PageLimit: limit,
	}
	if cursor != nil {
		params.Before = &cursor.CreatedAt
		params.BeforeID = &cursor.ID
	}

	row, err := r.ListNotifications(ctx, params)
	if err != nil {
		return nil, 0, erring.Wrap(err, "notificationrepo.List")
	}

	items, err := toNotifications(row.Items)

	return items, row.UnreadCount, err
}

func (r *NotificationRepository) UnreadCount(ctx context.Context, userID string) (int, error) {
	count, err := r.UnreadNotificationCount(ctx, userID)

	return count, erring.Wrap(err, "notificationrepo.UnreadCount")
}
