//go:build integration

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/config"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/testinfra"
)

type reportGenerator struct{}

func (reportGenerator) Generate(ctx context.Context, _ string, tools report.Toolset, progress report.Progress) (report.Content, error) {
	id, _ := ctxkey.GetRequestID(ctx)
	transport, _ := ctxkey.GetTransport(ctx)
	user, authenticated := ctxkey.GetUser(ctx)

	if !authenticated || !identifier.IsUUID(user.ID) || transport != "redis-stream" || !identifier.IsUUID(id) {
		return report.Content{}, errors.New("worker job context was lost")
	}

	progress("análise", "Conferindo contexto")
	progress("revisão", "Conferindo evidências")

	ids := tools.SourceIDs()

	return report.Content{Title: "Relatório de teste", Summary: "Evidências consultadas", Metrics: []report.Metric{}, Sections: []report.Section{}, Charts: []report.Chart{}, Tables: []report.Table{}, Recommendations: []string{}, MissingData: []string{}, SourceIDs: ids}, nil
}

func TestSecuredAPIRefreshMCPReportsAndWebSocket(t *testing.T) {
	t.Parallel()

	databaseURL := testinfra.Postgres(t)

	ctx := t.Context()

	db, err := postgres.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(db.Close)

	if err = db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	application, err := app.New(db)
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{Auth: config.Auth{JWTSecret: "test-only-secret-with-at-least-32-bytes", AccessTTL: 5 * time.Minute, RefreshTTL: time.Hour, AdminName: "Admin teste", AdminEmail: "integration-admin@example.com", AdminPassword: "integration-admin-password", AllowedOrigins: "http://localhost:5173"}, OpenAI: config.OpenAI{APIKey: "test-fake-key", Timeout: 10 * time.Second}}
	if err = application.Configure(ctx, cfg); err != nil {
		t.Fatal(err)
	}

	application.Reports.Generator = reportGenerator{}

	apiServer := api.New(application, cfg)
	defer apiServer.Close()

	server := httptest.NewServer(apiServer.Handler)
	defer server.Close()

	newClient := func() *http.Client { return testClient(t) }
	admin, member := newClient(), newClient()
	call := func(client *http.Client, method, path string, value any, status int) []byte {
		return callAPI(t, client, server.URL, method, path, value, status)
	}

	call(newClient(), "GET", "/api/v1/demands", nil, 401)
	call(newClient(), "GET", "/api/v1/releases", nil, 401)
	call(newClient(), "GET", "/api/v1/board", nil, 401)

	login := func(client *http.Client, email, password string) {
		call(client, "POST", "/api/v1/auth/login", map[string]string{"email": email, "password": password}, 200)
	}
	login(admin, cfg.Auth.AdminEmail, cfg.Auth.AdminPassword)

	apiURL, _ := url.Parse(server.URL + "/api/v1/auth")
	oldRefresh := refreshToken(t, admin, apiURL)

	for _, path := range []string{"/api/v1/auth/me", "/api/v1/metrics", "/api/v1/reports/demands.csv"} {
		call(admin, "GET", path, nil, 200)
	}

	assertCSRF(t, admin, server.URL)

	call(admin, "POST", "/api/v1/auth/refresh", map[string]any{}, 200)

	replay := newClient()
	replay.Jar.SetCookies(apiURL, []*http.Cookie{{Name: "sf_refresh", Value: oldRefresh, Path: "/api/v1/auth"}})
	call(replay, "POST", "/api/v1/auth/refresh", map[string]any{}, 401)
	call(admin, "GET", "/api/v1/auth/me", nil, 401)
	login(admin, cfg.Auth.AdminEmail, cfg.Auth.AdminPassword)

	id, _ := identifier.NewUUID()
	email := "member-" + id + "@example.com"
	userJSON := call(admin, "POST", "/api/v1/users", map[string]string{"name": "Usuário teste", "email": email, "password": "member-test-password"}, 201)

	var memberUser entity.User

	if err = json.Unmarshal(userJSON, &memberUser); err != nil {
		t.Fatal(err)
	}

	if memberUser.Role != "member" || bytes.Contains(userJSON, []byte("password")) {
		t.Fatal("unsafe user response")
	}

	call(admin, "POST", "/api/v1/users", map[string]string{"name": "Outro admin", "email": "admin2@example.com", "password": "member-test-password", "role": "admin"}, 400)
	login(member, email, "member-test-password")
	assertReleaseEndpoints(t, member, server.URL, db)
	assertCatalogEndpoints(t, admin, member, server.URL, memberUser.ID)
	assertNotificationEndpoints(t, admin, member, server.URL, memberUser.ID, db)
	call(member, "GET", "/api/v1/users", nil, 403)
	call(member, "POST", "/api/v1/users", map[string]string{"name": "Forbidden", "email": "forbidden@example.com", "password": "member-test-password"}, 403)
	call(member, "POST", "/api/v1/mcp", map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"protocolVersion": "2025-06-18", "capabilities": map[string]any{}, "clientInfo": map[string]string{"name": "integration", "version": "1"}}}, 200)

	list := call(member, "POST", "/api/v1/mcp", map[string]any{"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": map[string]any{}}, 200)
	if !bytes.Contains(list, []byte("get_demand_context")) {
		t.Fatal("MCP missing tool")
	}

	call(member, "POST", "/api/v1/mcp", map[string]any{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": map[string]any{"name": "get_metrics", "arguments": map[string]any{}}}, 200)

	wsHeaders := http.Header{"Origin": []string{"http://localhost:5173"}}
	rootURL, _ := url.Parse(server.URL + "/api/v1/ws")
	cookieValues := []string{}

	for _, c := range member.Jar.Cookies(rootURL) {
		cookieValues = append(cookieValues, c.Name+"="+c.Value)
	}

	wsHeaders.Set("Cookie", strings.Join(cookieValues, "; "))

	socket, upgradeResponse, err := websocket.DefaultDialer.DialContext(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/api/v1/ws", wsHeaders)
	if err != nil {
		t.Fatal(err)
	}

	if upgradeResponse != nil {
		defer upgradeResponse.Body.Close()
	}

	defer socket.Close()

	if deadlineErr := socket.SetReadDeadline(time.Now().Add(10 * time.Second)); deadlineErr != nil {
		t.Fatal(deadlineErr)
	}

	readEvent := func(expected string) { readSocketEvent(t, socket, expected) }

	readEvent("connected")

	created := call(admin, "POST", "/api/v1/demands", map[string]any{"title": "Demanda integração autenticada", "kind": "improvement", "priority": "medium", "description": "Contexto técnico", "acceptance_criteria": "Campo obrigatório validado"}, 201)

	var item entity.Demand

	if decodeErr := json.Unmarshal(created, &item); decodeErr != nil {
		t.Fatal(decodeErr)
	}

	readEvent("workspace.updated")

	for _, status := range []string{"in_progress", "in_qa"} {
		call(member, "POST", "/api/v1/demands/"+item.ID+"/move", map[string]string{"status": status, "actor": "FORGED"}, 200)
	}

	call(member, "POST", "/api/v1/demands/"+item.ID+"/tests", map[string]string{"title": "Cenário válido", "result": "passed", "author": "FORGED"}, 201)
	call(member, "POST", "/api/v1/demands/"+item.ID+"/move", map[string]string{"status": "done", "actor": "FORGED"}, 200)

	detailJSON := call(member, "GET", "/api/v1/demands/"+item.ID, nil, 200)
	if bytes.Contains(detailJSON, []byte("FORGED")) || !bytes.Contains(detailJSON, []byte(memberUser.ID)) {
		t.Fatal("audit identity was forged")
	}

	assertAsyncReportEndpoints(t, application, admin, member, server.URL, socket, item)

	assertPasswordEndpoints(t, admin, member, server.URL, db, memberUser)

	call(member, "POST", "/api/v1/auth/logout", map[string]any{}, 200)
	call(member, "GET", "/api/v1/demands", nil, 401)

	var admins int
	if err = db.Pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE role='admin'`).Scan(&admins); err != nil || admins != 1 {
		t.Fatalf("single admin: %d %v", admins, err)
	}
}

func testClient(t *testing.T) *http.Client {
	t.Helper()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	return &http.Client{Jar: jar}
}

func callAPI(t *testing.T, client *http.Client, baseURL, method, path string, value any, status int) []byte {
	t.Helper()

	var body io.Reader

	if value != nil {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}

		body = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(t.Context(), method, baseURL+path, body)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Squadflow-Client", "web")
	req.Header.Set("X-Request-ID", "integration-http")

	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer response.Body.Close()

	if response.Header.Get("X-Request-ID") != "integration-http" {
		t.Fatal("HTTP request correlation was lost")
	}

	out, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != status {
		t.Fatalf("%s %s: status %d want %d body=%s", method, path, response.StatusCode, status, out)
	}

	return out
}

func readSocketEvent(t *testing.T, conn *websocket.Conn, expected string) map[string]any {
	t.Helper()

	for range 20 {
		var event map[string]any
		if err := conn.ReadJSON(&event); err != nil {
			t.Fatal(err)
		}

		if event["type"] == expected {
			return event
		}
	}

	t.Fatalf("missing %s", expected)

	return nil
}

func assertCSRF(t *testing.T, admin *http.Client, baseURL string) {
	t.Helper()

	for _, headers := range []map[string]string{{"Content-Type": "application/json", "X-SquadFlow-Client": "web", "Origin": "https://evil.example"}, {"Content-Type": "application/json"}} {
		req, requestErr := http.NewRequestWithContext(t.Context(), http.MethodPost, baseURL+"/api/v1/auth/refresh", strings.NewReader("{}"))
		if requestErr != nil {
			t.Fatal(requestErr)
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}

		res, requestErr := admin.Do(req)
		if requestErr != nil {
			t.Fatal(requestErr)
		}

		res.Body.Close()

		if res.StatusCode != http.StatusForbidden {
			t.Fatalf("CSRF: %d", res.StatusCode)
		}
	}
}

func refreshToken(t *testing.T, client *http.Client, apiURL *url.URL) string {
	t.Helper()

	var refresh, access string

	for _, cookie := range client.Jar.Cookies(apiURL) {
		if cookie.Name == "sf_refresh" {
			refresh = cookie.Value
		}

		if cookie.Name == "sf_access" {
			access = cookie.Value
		}
	}

	if refresh == "" || access == "" {
		t.Fatal("missing cookie tokens")
	}

	return refresh
}

func assertReleaseEndpoints(t *testing.T, client *http.Client, serverURL string, db *postgres.Client) {
	t.Helper()

	marker, err := identifier.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	call := func(method, path string, input any, status int) []byte {
		return callAPI(t, client, serverURL, method, path, input, status)
	}
	raw := call("POST", "/api/v1/releases", map[string]string{"name": "API-" + marker}, 201)

	var version entity.Release

	if err = json.Unmarshal(raw, &version); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		ctx := context.WithoutCancel(t.Context())
		if _, cleanupErr := db.Pool.Exec(ctx, `DELETE FROM work_items WHERE release_id=$1`, version.ID); cleanupErr != nil {
			t.Error(cleanupErr)
		}

		if _, cleanupErr := db.Pool.Exec(ctx, `DELETE FROM releases WHERE id=$1`, version.ID); cleanupErr != nil {
			t.Error(cleanupErr)
		}
	})
	call("PUT", "/api/v1/releases/"+version.ID, map[string]string{"name": version.Name, "description": "Notas da versão"}, 200)
	call("GET", "/api/v1/releases", nil, 200)
	call("GET", "/api/v1/board?limit=invalid", nil, 400)
	call("GET", "/api/v1/board?limit=101", nil, 400)
	call("GET", "/api/v1/board?offset=-1", nil, 400)
	call("POST", "/api/v1/releases/"+version.ID+"/archive", map[string]any{}, 400)
	call("POST", "/api/v1/releases/not-a-uuid/archive", map[string]any{"archived": true}, 400)
	itemRaw := call("POST", "/api/v1/demands", map[string]string{"title": "Demanda da versão API " + marker, "kind": "improvement", "priority": "medium", "release_id": version.ID}, 201)

	var item entity.Demand

	if err = json.Unmarshal(itemRaw, &item); err != nil {
		t.Fatal(err)
	}

	if item.ReleaseID != version.ID || item.ReleaseName != version.Name {
		t.Fatal("API ignored release_id")
	}

	call("POST", "/api/v1/releases/"+version.ID+"/archive", map[string]bool{"archived": true}, 200)
	activeRaw := call("GET", "/api/v1/board?release_id="+version.ID, nil, 200)

	var active demand.Board

	if err = json.Unmarshal(activeRaw, &active); err != nil {
		t.Fatal(err)
	}

	if len(active.Columns) != 5 || active.Columns[0].Total != 0 {
		t.Fatal("archived release is still active")
	}

	archivedRaw := call("GET", "/api/v1/board?archive=archived&release_id="+version.ID, nil, 200)
	if !bytes.Contains(archivedRaw, []byte(item.ID)) {
		t.Fatal("archived demand missing")
	}

	if !bytes.Contains(call("GET", "/api/v1/reports/demands.csv", nil, 200), []byte(version.Name)) {
		t.Fatal("CSV lost version")
	}

	call("GET", "/api/v1/demands/"+item.ID, nil, 200)
	call("POST", "/api/v1/releases/"+version.ID+"/archive", map[string]bool{"archived": false}, 200)
}

func assertAsyncReportEndpoints(t *testing.T, application *app.App, admin, member *http.Client, serverURL string, socket *websocket.Conn, item entity.Demand) {
	t.Helper()

	call := func(client *http.Client, method, path string, value any, status int) []byte {
		return callAPI(t, client, serverURL, method, path, value, status)
	}
	readEvent := func(expected string) map[string]any { return readSocketEvent(t, socket, expected) }
	generatedJSON := call(member, "POST", "/api/v1/reports/generate", map[string]string{"prompt": "Avalie a demanda concluída", "demand_id": item.ID}, 202)

	var queued report.Job
	if decodeErr := json.Unmarshal(generatedJSON, &queued); decodeErr != nil {
		t.Fatal(decodeErr)
	}

	if queued.Status != report.JobQueued {
		t.Fatalf("request did not return queued job: %+v", queued)
	}

	call(admin, "GET", "/api/v1/reports/jobs/"+queued.ID, nil, 404)

	worker := report.Worker{Repository: application.ReportJobs.Repository, Reports: application.Reports, Timeout: 10 * time.Second}
	if processErr := worker.Process(t.Context(), report.QueueEvent{ID: queued.ID, JobID: queued.ID, Type: report.EventRequested}); processErr != nil {
		t.Fatal(processErr)
	}

	generatedJSON = call(member, "GET", "/api/v1/reports/"+queued.ID, nil, 200)

	var generated report.Report

	if decodeErr := json.Unmarshal(generatedJSON, &generated); decodeErr != nil {
		t.Fatal(decodeErr)
	}

	if len(generated.SourceData) == 0 || !bytes.Contains(generated.SourceData, []byte("Cenário válido")) {
		t.Fatal("report lacks full QA snapshot")
	}

	call(admin, "GET", "/api/v1/reports/"+generated.ID, nil, 404)
	call(member, "GET", "/api/v1/reports/"+generated.ID, nil, 200)

	if err := socket.WriteJSON(map[string]any{"type": "report.generate", "request_id": "test-request", "request": map[string]string{"prompt": "Avalie os testes e o histórico", "demand_id": item.ID}}); err != nil {
		t.Fatal(err)
	}

	ready := readEvent("report.queued")
	if ready["request_id"] != "test-request" {
		t.Fatal("correlation missing")
	}

	var socketJob report.Job

	socketData, encodeErr := json.Marshal(ready["data"])
	if encodeErr != nil {
		t.Fatal(encodeErr)
	}

	if decodeErr := json.Unmarshal(socketData, &socketJob); decodeErr != nil {
		t.Fatal(decodeErr)
	}

	_ = socket.Close()

	if processErr := worker.Process(t.Context(), report.QueueEvent{ID: socketJob.ID, JobID: socketJob.ID, Type: report.EventRequested}); processErr != nil {
		t.Fatal(processErr)
	}

	call(member, "GET", "/api/v1/reports/"+socketJob.ID, nil, 200)
	call(admin, "GET", "/api/v1/reports/"+socketJob.ID+"/html", nil, 404)
	assertReportHTMLDownload(t, member, serverURL, socketJob.ID)
}

func assertReportHTMLDownload(t *testing.T, client *http.Client, serverURL, id string) {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"/api/v1/reports/"+id+"/html?download=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}

	defer response.Body.Close()

	document, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusOK || !strings.Contains(response.Header.Get("Content-Type"), "text/html") || !strings.Contains(response.Header.Get("Content-Disposition"), "attachment") || !strings.Contains(response.Header.Get("Content-Security-Policy"), "script-src 'none'") {
		t.Fatalf("invalid HTML download headers: %+v", response.Header)
	}

	if !bytes.Contains(document, []byte("Content-Security-Policy")) || !bytes.Contains(document, []byte("Relatório de teste")) {
		t.Fatal("HTML export missing report or offline security policy")
	}
}
