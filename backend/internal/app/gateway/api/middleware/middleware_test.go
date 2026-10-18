package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/auth"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/middleware"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/realtime"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
)

func TestRequestContextBridgesChiAndIsolatesRequests(t *testing.T) {
	t.Parallel()

	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID, middleware.RequestContext)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id, ok := ctxkey.GetRequestID(ctx)

		if !ok || id == "" || id != chimiddleware.GetReqID(ctx) || id != w.Header().Get("X-Request-ID") {
			t.Error("request ID missing or inconsistent")
		}

		transport, _ := ctxkey.GetTransport(ctx)
		if transport != "http" {
			t.Errorf("unexpected transport: %s", transport)
		}

		for _, metadata := range []struct {
			header string
			get    func(context.Context) (string, bool)
		}{{"Authorization", ctxkey.GetAuthorizationHeader}, {"Idempotency-Key", ctxkey.GetIdempotencyKey}} {
			value, exists := metadata.get(ctx)
			expected := r.Header.Get(metadata.header)

			if value != expected || exists != (expected != "") {
				t.Errorf("%s metadata leaked or missing", metadata.header)
			}
		}

		w.WriteHeader(http.StatusNoContent)
	})

	provided := httptest.NewRequest(http.MethodGet, "/", nil)
	provided.Header.Set("X-Request-ID", "client-request")
	provided.Header.Set("Authorization", "Bearer test-token")
	provided.Header.Set("Idempotency-Key", "mutation-key")

	recorded := httptest.NewRecorder()
	router.ServeHTTP(recorded, provided)

	if recorded.Header().Get("X-Request-ID") != "client-request" {
		t.Fatal("client request ID was lost")
	}

	ids := map[string]bool{"client-request": true}

	for range 2 {
		recorded = httptest.NewRecorder()
		router.ServeHTTP(recorded, httptest.NewRequest(http.MethodGet, "/", nil))

		id := recorded.Header().Get("X-Request-ID")
		if ids[id] {
			t.Fatal("generated request ID reused across requests")
		}

		ids[id] = true
	}
}

func TestRequireAuthPreservesCredentialPrecedenceAndRejectsFailures(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, header, cookie, token string
		err                         error
		status                      int
	}{
		{name: "cookie", cookie: "cookie-token", token: "cookie-token", status: 204},
		{name: "bearer takes precedence", header: "Bearer bearer-token", cookie: "cookie-token", token: "bearer-token", status: 204},
		{name: "invalid bearer never falls back", header: "Bearer expired", cookie: "cookie-token", token: "expired", err: auth.ErrUnauthorized, status: 401},
		{name: "absent credentials", err: auth.ErrUnauthorized, status: 401},
		{name: "authentication failure is hidden", cookie: "cookie-token", token: "cookie-token", err: errors.New("private database failure"), status: 500},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := NewMockAuthenticator(gomock.NewController(t))
			service.EXPECT().Authenticate(gomock.Any(), test.token).Return(auth.Principal{User: entity.User{ID: "authenticated-user", Role: "member"}}, test.err)

			reached := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reached = true

				user, ok := ctxkey.GetUser(r.Context())
				if !ok || user.ID != "authenticated-user" || user.Role != "member" {
					t.Error("authenticated identity not available to handler")
				}

				w.WriteHeader(http.StatusNoContent)
			})
			router := chi.NewRouter()
			router.Use(chimiddleware.RequestID, middleware.RequestContext, middleware.RequireAuth(service))
			router.Get("/", next)

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set("Authorization", test.header)

			if test.cookie != "" {
				request.AddCookie(&http.Cookie{Name: "sf_access", Value: test.cookie})
			}

			recorded := httptest.NewRecorder()
			router.ServeHTTP(recorded, request)

			if recorded.Code != test.status || reached != (test.err == nil) {
				t.Fatalf("unexpected authentication result: status=%d handler=%t", recorded.Code, reached)
			}

			if test.err != nil && (!strings.Contains(recorded.Header().Get("Content-Type"), "application/json") || strings.Contains(recorded.Body.String(), "private")) {
				t.Fatal("authentication error contract changed or leaked details")
			}
		})
	}
}

func TestOriginProtectionPreservesCookieAndBearerRules(t *testing.T) {
	t.Parallel()

	policy := middleware.NewOriginPolicy(" https://app.example.com, https://admin.example.com ")
	if policy.Allowed("") || policy.Allowed("https://app.example.com.evil") || !policy.Allowed("https://admin.example.com") {
		t.Fatal("origin allowlist is not exact")
	}

	for _, test := range []struct {
		name, method, origin, content, client string
		status                                int
	}{
		{"allowed read", "GET", "https://app.example.com", "", "", 204},
		{"foreign read", "GET", "https://evil.example.com", "", "", 403},
		{"allowed mutation", "POST", "https://app.example.com", "application/json; charset=utf-8", "web", 204},
		{"missing client header", "POST", "https://app.example.com", "application/json", "", 403},
		{"foreign mutation", "PUT", "https://evil.example.com", "application/json", "web", 403},
		{"form mutation", "POST", "https://app.example.com", "application/x-www-form-urlencoded", "web", 403},
		{"bearer without origin", "POST", "", "application/json", "web", 204},
		{"bearer still requires csrf header", "POST", "", "application/json", "", 403},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
			request := httptest.NewRequest(test.method, "/", nil)
			request.Header.Set("Origin", test.origin)
			request.Header.Set("Content-Type", test.content)
			request.Header.Set("X-Squadflow-Client", test.client)
			request.Header.Set("Authorization", "Bearer client-token")

			recorded := httptest.NewRecorder()
			policy.Protect(next).ServeHTTP(recorded, request)

			if recorded.Code != test.status {
				t.Fatalf("expected %d, received %d", test.status, recorded.Code)
			}
		})
	}
}

func TestWorkspaceUpdatesOnlyFollowCommittedSuccessfulMutations(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, method string
		statuses     []int
		published    bool
	}{
		{"created", "POST", []int{201}, true},
		{"empty success", "PUT", nil, true},
		{"read", "GET", []int{200}, false},
		{"conflict", "POST", []int{409}, false},
		{"redirect", "POST", []int{302}, false},
		{"first status failed", "PUT", []int{500, 200}, false},
		{"first status succeeded", "POST", []int{201, 400}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			events := NewMockPublisher(gomock.NewController(t))
			if test.published {
				events.EXPECT().Publish("", gomock.Cond(func(event realtime.Event) bool { return event.Type == "workspace.updated" }))
			}

			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				for _, status := range test.statuses {
					w.WriteHeader(status)
				}
			})
			request := httptest.NewRequest(test.method, "/api/v1/demands", nil)
			middleware.WorkspaceUpdates(events)(next).ServeHTTP(httptest.NewRecorder(), request)
		})
	}
}

func TestMiddlewareChainRecoversPanicsWithoutPublishingUpdates(t *testing.T) {
	t.Parallel()

	events := NewMockPublisher(gomock.NewController(t))
	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID, middleware.RequestContext, middleware.RequestLogger, chimiddleware.Recoverer, chimiddleware.NoCache)
	router.With(middleware.WorkspaceUpdates(events)).Post("/mutation", func(_ http.ResponseWriter, _ *http.Request) {
		panic("handler failed before committing")
	})
	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })

	recorded := httptest.NewRecorder()
	router.ServeHTTP(recorded, httptest.NewRequest(http.MethodPost, "/mutation", nil))

	if recorded.Code != http.StatusInternalServerError || strings.Contains(recorded.Body.String(), "handler failed") {
		t.Fatal("panic was not recovered safely or triggered a workspace update")
	}

	if recorded.Header().Get("X-Request-ID") == "" || !strings.Contains(recorded.Header().Get("Cache-Control"), "no-store") {
		t.Fatal("panic response lost the request metadata or cache policy")
	}

	recorded = httptest.NewRecorder()
	router.ServeHTTP(recorded, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if recorded.Code != http.StatusNoContent {
		t.Fatal("router did not serve requests after recovering")
	}
}
