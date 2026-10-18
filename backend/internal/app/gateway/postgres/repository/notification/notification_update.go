package notificationrepo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *NotificationRepository) Read(ctx context.Context, userID, id string) (time.Time, error) {
	readAt, err := r.ReadNotification(ctx, sqlc.ReadNotificationParams{ID: id, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return readAt, erring.ErrNotFound
	}

	return readAt, erring.Wrap(err, "notificationrepo.Read")
}

func (r *NotificationRepository) ReadAll(ctx context.Context, userID string) (time.Time, error) {
	readAt, err := r.ReadAllNotifications(ctx, userID)

	return readAt, erring.Wrap(err, "notificationrepo.ReadAll")
}
