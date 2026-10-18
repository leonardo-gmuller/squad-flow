package notification

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

func TestAdminAndInvalidPrincipalsNeverAccessInbox(t *testing.T) {
	t.Parallel()

	for _, user := range []entity.User{{ID: "11111111-1111-4111-8111-111111111111", Role: "admin"}, {Role: "member"}, {ID: "invalid", Role: "member"}} {
		repo := NewMockRepository(gomock.NewController(t))
		service := New(repo)
		_, listErr := service.List(t.Context(), user, "all", 20, "")
		_, countErr := service.UnreadCount(t.Context(), user)
		_, readErr := service.Read(t.Context(), user, "11111111-1111-4111-8111-111111111112")
		_, allErr := service.ReadAll(t.Context(), user)

		for _, err := range []error{listErr, countErr, readErr, allErr} {
			if !errors.Is(err, erring.ErrForbidden) {
				t.Fatal("inbox accessible to unsupported principal", err)
			}
		}
	}
}

func TestCursorContinuesAfterLastVisibleItemAndKeepsAuthenticatedOwner(t *testing.T) {
	t.Parallel()

	user := entity.User{ID: "11111111-1111-4111-8111-111111111111", Role: "member"}
	now := time.Now().UTC()
	items := []entity.Notification{{ID: "11111111-1111-4111-8111-111111111112", CreatedAt: now}, {ID: "11111111-1111-4111-8111-111111111113", CreatedAt: now.Add(-time.Second)}}
	repo := NewMockRepository(gomock.NewController(t))
	gomock.InOrder(
		repo.EXPECT().List(gomock.Any(), user.ID, true, 2, nil).Return(items, 7, nil),
		repo.EXPECT().List(gomock.Any(), user.ID, true, 2, &Cursor{ID: items[0].ID, CreatedAt: now}).Return(items[1:], 7, nil),
	)

	service := New(repo)

	page, err := service.List(t.Context(), user, "unread", 1, "")
	if err != nil || len(page.Items) != 1 || page.UnreadCount != 7 || page.NextCursor == "" {
		t.Fatal("invalid first page", err)
	}

	raw, err := base64.RawURLEncoding.DecodeString(page.NextCursor)
	if err != nil {
		t.Fatal(err)
	}

	var cursor Cursor
	if err = json.Unmarshal(raw, &cursor); err != nil || cursor.ID != page.Items[0].ID || !cursor.CreatedAt.Equal(now) {
		t.Fatal("cursor skipped the extra item", err)
	}

	if _, err = service.List(t.Context(), user, "unread", 1, page.NextCursor); err != nil {
		t.Fatal("cursor or owner was not preserved", err)
	}
}

func TestRejectsInvalidPaginationWithoutQuerying(t *testing.T) {
	t.Parallel()

	user := entity.User{ID: "11111111-1111-4111-8111-111111111111", Role: "member"}

	for _, input := range []struct {
		filter string
		limit  int
		cursor string
	}{{"other", 20, ""}, {"all", 51, ""}, {"all", -1, ""}, {"all", 20, "invalid"}, {"all", 20, base64.RawURLEncoding.EncodeToString([]byte(`{"id":"invalid","at":"2026-10-17T00:00:00Z"}`))}} {
		repo := NewMockRepository(gomock.NewController(t))
		if _, err := New(repo).List(t.Context(), user, input.filter, input.limit, input.cursor); !errors.Is(err, erring.ErrInvalidInput) {
			t.Fatal("invalid pagination reached repository", err)
		}
	}
}
