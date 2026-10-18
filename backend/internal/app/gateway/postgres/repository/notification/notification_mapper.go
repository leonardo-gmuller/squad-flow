package notificationrepo

import (
	"encoding/json"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

func toNotifications(raw []byte) ([]entity.Notification, error) {
	items := []entity.Notification{}
	err := json.Unmarshal(raw, &items)

	return items, erring.Wrap(err, "notificationrepo.toNotifications")
}
