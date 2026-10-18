//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/notification"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres"
	demandrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/demand"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

func assertNotificationEndpoints(t *testing.T, admin, member *http.Client, baseURL, userID string, db *postgres.Client) {
	t.Helper()

	call := func(client *http.Client, method, path string, input any, status int) []byte {
		return callAPI(t, client, baseURL, method, path, input, status)
	}
	count := func(client *http.Client) int {
		return notificationUnreadCount(t, client, baseURL)
	}
	list := func(client *http.Client, query string) notification.Page {
		var page notification.Page
		if err := json.Unmarshal(call(client, "GET", "/api/v1/notifications"+query, nil, 200), &page); err != nil {
			t.Fatal(err)
		}

		return page
	}

	call(testClient(t), "GET", "/api/v1/notifications", nil, 401)
	call(admin, "GET", "/api/v1/notifications", nil, 403)
	call(admin, "GET", "/api/v1/notifications/unread-count", nil, 403)
	call(admin, "POST", "/api/v1/notifications/read-all", map[string]any{}, 403)
	call(member, "POST", "/api/v1/notifications/read-all", map[string]any{}, 200)

	if count(member) != 0 {
		t.Fatal("mark all did not clear member notifications")
	}

	id, _ := identifier.NewUUID()
	email := "notification-" + id + "@example.test"

	var other entity.User

	if err := json.Unmarshal(call(admin, "POST", "/api/v1/users", map[string]string{"name": "Outro destinatário", "email": email, "password": "notification-test-password"}, 201), &other); err != nil {
		t.Fatal(err)
	}

	otherClient := testClient(t)
	call(otherClient, "POST", "/api/v1/auth/login", map[string]string{"email": email, "password": "notification-test-password"}, 200)

	var item entity.Demand
	if err := json.Unmarshal(call(admin, "POST", "/api/v1/demands", map[string]string{"title": "Entrega para notificação privada", "kind": "improvement", "priority": "medium", "assignee_id": userID, "reviewer_id": userID}, 201), &item); err != nil {
		t.Fatal(err)
	}

	page := list(member, "?filter=unread")
	if count(member) != 1 || len(page.Items) != 1 || page.Items[0].Kind != "assigned" || page.Items[0].DemandID != item.ID {
		t.Fatal("assignment not deduplicated or not delivered")
	}

	first := page.Items[0]
	call(otherClient, "POST", "/api/v1/notifications/"+first.ID+"/read", map[string]any{}, 404)
	call(admin, "POST", "/api/v1/notifications/"+first.ID+"/read", map[string]any{}, 403)

	if count(member) != 1 || count(otherClient) != 0 {
		t.Fatal("another account changed private notifications")
	}

	if len(list(otherClient, "?user_id="+userID).Items) != 0 {
		t.Fatal("member can read another user's inbox")
	}

	raw := call(member, "POST", "/api/v1/notifications/"+first.ID+"/read", map[string]any{}, 200)
	if string(call(member, "POST", "/api/v1/notifications/"+first.ID+"/read", map[string]any{}, 200)) != string(raw) {
		t.Fatal("reading a notification is not idempotent")
	}

	if count(member) != 0 || len(list(member, "?filter=unread").Items) != 0 {
		t.Fatal("read state not persisted")
	}

	edit := map[string]any{"title": item.Title, "priority": "medium", "assignee_id": other.ID, "reviewer_id": userID}
	call(admin, "PUT", "/api/v1/demands/"+item.ID, edit, 200)

	if count(member) != 1 || count(otherClient) != 1 {
		t.Fatal("changed responsibilities were not delivered")
	}

	call(admin, "POST", "/api/v1/demands/"+item.ID+"/move", map[string]string{"status": "in_progress"}, 200)

	if count(member) != 2 || count(otherClient) != 2 {
		t.Fatal("status change missed a recipient")
	}

	call(otherClient, "POST", "/api/v1/demands/"+item.ID+"/move", map[string]string{"status": "in_qa"}, 200)

	if count(member) != 3 || count(otherClient) != 2 {
		t.Fatal("self action generated a notification")
	}

	call(member, "POST", "/api/v1/demands/"+item.ID+"/tests", map[string]string{"title": "Cenário aprovado", "result": "passed"}, 201)

	if count(otherClient) != 3 || count(member) != 3 {
		t.Fatal("QA approval not delivered to the other participant")
	}

	call(member, "POST", "/api/v1/demands/"+item.ID+"/tests", map[string]string{"title": "Cenário reprovado", "result": "failed", "actual": "Revisar comportamento"}, 201)

	if count(otherClient) != 4 {
		t.Fatal("QA rejection duplicated or lost the return notification")
	}

	actor := assertNotificationAdminExcluded(t, admin, member, baseURL, db)

	assertNotificationPagination(t, member, baseURL, userID, item.ID, db)

	for _, query := range []string{"?limit=0", "?limit=-1", "?limit=51", "?filter=invalid", "?before=invalid"} {
		call(member, "GET", "/api/v1/notifications"+query, nil, 400)
	}

	call(member, "POST", "/api/v1/notifications/invalid/read", map[string]any{}, 400)

	before := count(otherClient)
	assertNotificationRollback(t, otherClient, baseURL, actor, other.ID, db, before)
	call(member, "POST", "/api/v1/notifications/read-all", map[string]any{}, 200)

	if count(member) != 0 || count(otherClient) != before {
		t.Fatal("mark all modified another user's inbox")
	}
	// No user-controlled recipient parameter changes which inbox is read.
	if got := list(member, "?filter=unread&user_id="+other.ID); len(got.Items) != 0 {
		t.Fatal("recipient spoofing changed the inbox")
	}
}

func assertNotificationPagination(t *testing.T, member *http.Client, baseURL, userID, demandID string, db *postgres.Client) {
	t.Helper()

	list := func(query string) notification.Page {
		var page notification.Page
		if err := json.Unmarshal(callAPI(t, member, baseURL, "GET", "/api/v1/notifications"+query, nil, 200), &page); err != nil {
			t.Fatal(err)
		}

		return page
	}
	// Cursor pages remain stable when a newer notification arrives between pages.
	for index := range 24 {
		notificationID, _ := identifier.NewUUID()
		if _, err := db.Pool.Exec(t.Context(), `INSERT INTO notifications(id,user_id,demand_id,kind,title,message,created_at) VALUES($1,$2,$3,'assigned','Paginação','Notificação de teste',$4)`, notificationID, userID, demandID, time.Date(2026, 1, 1, 0, 0, index, 0, time.UTC)); err != nil {
			t.Fatal(err)
		}
	}

	seen := map[string]bool{}
	page := list("?limit=2")

	for _, value := range page.Items {
		seen[value.ID] = true
	}

	newerID, _ := identifier.NewUUID()
	if _, err := db.Pool.Exec(t.Context(), `INSERT INTO notifications(id,user_id,demand_id,kind,title,message) VALUES($1,$2,$3,'assigned','Chegou depois','Nova notificação')`, newerID, userID, demandID); err != nil {
		t.Fatal(err)
	}

	for page.NextCursor != "" {
		page = list("?limit=2&before=" + page.NextCursor)
		for _, value := range page.Items {
			if seen[value.ID] || value.ID == newerID {
				t.Fatal("cursor duplicated or reordered notifications")
			}

			seen[value.ID] = true
		}
	}

	if len(seen) < 24 {
		t.Fatal("pagination lost notifications")
	}
}

func notificationUnreadCount(t *testing.T, member *http.Client, baseURL string) int {
	t.Helper()

	var result struct {
		Count int `json:"count"`
	}

	if err := json.Unmarshal(callAPI(t, member, baseURL, "GET", "/api/v1/notifications/unread-count", nil, 200), &result); err != nil {
		t.Fatal(err)
	}

	return result.Count
}

func assertNotificationRollback(t *testing.T, member *http.Client, baseURL string, actor entity.User, userID string, db *postgres.Client, before int) {
	t.Helper()
	// Failure after the notification INSERT must roll back both the demand and its event.
	unit, err := postgres.NewUnitOfWork(db)
	if err != nil {
		t.Fatal(err)
	}

	rollback := errors.New("force transaction rollback")
	rollbackID, _ := identifier.NewUUID()

	err = unit.Do(ctxkey.PutUser(t.Context(), actor), func(ctx context.Context, repos demand.Repositories) error {
		value := entity.Demand{ID: rollbackID, Title: "Entrega que não deve persistir", Kind: entity.Improvement, Priority: "medium", Status: entity.Backlog, AssigneeID: userID, EventCodes: []string{}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
		if insertErr := repos.Demands().Insert(ctx, &value); insertErr != nil {
			return fmt.Errorf("insert rollback fixture: %w", insertErr)
		}

		return rollback
	})
	if !errors.Is(err, rollback) || notificationUnreadCount(t, member, baseURL) != before {
		t.Fatal("rolled-back action leaked a notification", err)
	}

	if _, err := demandrepo.NewDemandRepository(db.Pool).Detail(t.Context(), rollbackID); err == nil {
		t.Fatal("rolled-back demand persisted")
	}
}

func assertNotificationAdminExcluded(t *testing.T, admin, member *http.Client, baseURL string, db *postgres.Client) entity.User {
	t.Helper()

	call := func(client *http.Client, method, path string, input any, status int) []byte {
		return callAPI(t, client, baseURL, method, path, input, status)
	}
	// Admins are excluded even when a member assigns a demand to an admin.
	var identity struct {
		User entity.User `json:"user"`
	}

	if err := json.Unmarshal(call(admin, "GET", "/api/v1/auth/me", nil, 200), &identity); err != nil {
		t.Fatal(err)
	}

	call(member, "POST", "/api/v1/demands", map[string]string{"title": "Entrega atribuída ao administrador", "kind": "improvement", "priority": "medium", "assignee_id": identity.User.ID}, 201)

	var admins int
	if err := db.Pool.QueryRow(t.Context(), `SELECT count(*) FROM notifications n JOIN users u ON u.id=n.user_id WHERE u.role='admin'`).Scan(&admins); err != nil || admins != 0 {
		t.Fatal("admin received notifications", err)
	}

	return identity.User
}
