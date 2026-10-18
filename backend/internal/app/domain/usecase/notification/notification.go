package notification

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

type Cursor struct {
	CreatedAt time.Time `json:"at"`
	ID        string    `json:"id"`
}
type Page struct {
	Items       []entity.Notification `json:"items"`
	UnreadCount int                   `json:"unread_count"`
	NextCursor  string                `json:"next_cursor"`
}
type Repository interface {
	List(ctx context.Context, userID string, unread bool, limit int, cursor *Cursor) ([]entity.Notification, int, error)
	UnreadCount(ctx context.Context, userID string) (int, error)
	Read(ctx context.Context, userID, id string) (time.Time, error)
	ReadAll(ctx context.Context, userID string) (time.Time, error)
}
type UseCase struct{ repo Repository }

func New(repo Repository) *UseCase { return &UseCase{repo: repo} }
func RequireMember(user entity.User) error {
	if user.Role != "member" || !identifier.IsUUID(user.ID) {
		return erring.ErrForbidden.WithMessage("a central de notificações está disponível apenas para usuários membros")
	}

	return nil
}

func (s *UseCase) List(ctx context.Context, user entity.User, filter string, limit int, before string) (Page, error) {
	if err := RequireMember(user); err != nil {
		return Page{}, err
	}

	if filter == "" {
		filter = "all"
	}

	if limit == 0 {
		limit = 20
	}

	if (filter != "all" && filter != "unread") || limit < 1 || limit > 50 || len(before) > 256 {
		return Page{}, erring.ErrInvalidInput.WithMessage("filtros de notificações inválidos")
	}

	var cursor *Cursor

	if before != "" {
		raw, err := base64.RawURLEncoding.DecodeString(before)

		var value Cursor

		if err != nil || json.Unmarshal(raw, &value) != nil || value.CreatedAt.IsZero() || !identifier.IsUUID(value.ID) {
			return Page{}, erring.ErrInvalidInput.WithMessage("página de notificações inválida")
		}

		cursor = &value
	}

	items, count, err := s.repo.List(ctx, user.ID, filter == "unread", limit+1, cursor)
	if err != nil {
		return Page{}, erring.Wrap(err, "notification.List")
	}

	page := Page{Items: items, UnreadCount: count}
	if len(items) > limit {
		page.Items = items[:limit]
		last := page.Items[len(page.Items)-1]

		raw, err := json.Marshal(Cursor{CreatedAt: last.CreatedAt, ID: last.ID})
		if err != nil {
			return Page{}, erring.Wrap(err, "notification.List.cursor")
		}

		page.NextCursor = base64.RawURLEncoding.EncodeToString(raw)
	}

	if page.Items == nil {
		page.Items = []entity.Notification{}
	}

	return page, nil
}

func (s *UseCase) UnreadCount(ctx context.Context, user entity.User) (int, error) {
	if err := RequireMember(user); err != nil {
		return 0, err
	}

	count, err := s.repo.UnreadCount(ctx, user.ID)

	return count, erring.Wrap(err, "notification.UnreadCount")
}

func (s *UseCase) Read(ctx context.Context, user entity.User, id string) (time.Time, error) {
	if err := RequireMember(user); err != nil {
		return time.Time{}, err
	}

	if !identifier.IsUUID(id) {
		return time.Time{}, erring.ErrInvalidInput.WithMessage("notificação inválida")
	}

	readAt, err := s.repo.Read(ctx, user.ID, id)

	return readAt, erring.Wrap(err, "notification.Read")
}

func (s *UseCase) ReadAll(ctx context.Context, user entity.User) (time.Time, error) {
	if err := RequireMember(user); err != nil {
		return time.Time{}, err
	}

	readAt, err := s.repo.ReadAll(ctx, user.ID)

	return readAt, erring.Wrap(err, "notification.ReadAll")
}
