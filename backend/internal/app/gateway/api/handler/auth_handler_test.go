package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/go-chi/chi/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/config"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/auth"
	authmock "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/auth/mocks"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/response"
)

func TestLoginCredentialsHaveSamePublicErrorForUnknownAccountAndWrongPassword(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name        string
		lookupError error
		hash        string
	}{{"wrong password", nil, "test-hash"}, {"unknown account", auth.ErrUnauthorized, "dummy-hash"}} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			repo := NewMockRepository(ctrl)
			passwords := NewMockPasswords(ctrl)
			tokens := NewMockTokens(ctrl)

			passwords.EXPECT().Hash("invalid-login-timing-placeholder").Return("dummy-hash", nil)
			repo.EXPECT().ByEmail(gomock.Any(), "test@example.test").Return(entity.User{PasswordHash: "test-hash"}, test.lookupError)
			passwords.EXPECT().Match(test.hash, "wrong-password").Return(false)
			service := auth.New(repo, nil, passwords, tokens, time.Minute, time.Hour)
			router := chi.NewRouter()
			NewAuth(service, config.Auth{}).RegisterPublic(router)

			request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"test@example.test","password":"wrong-password"}`))
			recorded := httptest.NewRecorder()
			router.ServeHTTP(recorded, request)

			var body response.Error
			if err := json.Unmarshal(recorded.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}

			if recorded.Code != http.StatusUnauthorized || body.Type != "srn:error:unauthorized" || body.Code != "auth:invalid-credentials" || body.Message != "Email ou senha inválidos." || len(recorded.Header().Values("Set-Cookie")) != 0 {
				t.Fatalf("incorrect login contract: status=%d body=%s", recorded.Code, recorded.Body.String())
			}
		})
	}
}

func TestLoginDoesNotMisclassifyRepositoryFailureAsInvalidCredentials(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	passwords := NewMockPasswords(ctrl)
	passwords.EXPECT().Hash("invalid-login-timing-placeholder").Return("dummy-hash", nil)
	repo.EXPECT().ByEmail(gomock.Any(), "test@example.test").Return(entity.User{}, errors.New("private repository failure"))
	service := auth.New(repo, nil, passwords, NewMockTokens(ctrl), time.Minute, time.Hour)
	router := chi.NewRouter()
	NewAuth(service, config.Auth{}).RegisterPublic(router)

	recorded := httptest.NewRecorder()
	router.ServeHTTP(recorded, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"test@example.test","password":"wrong-password"}`)))

	if recorded.Code != http.StatusInternalServerError || strings.Contains(recorded.Body.String(), "private") || strings.Contains(recorded.Body.String(), "auth:invalid-credentials") {
		t.Fatalf("internal error was exposed or misclassified: %s", recorded.Body.String())
	}
}

func TestLoginHandlerAcceptsUseCaseMockAndSetsSessionCookies(t *testing.T) {
	t.Parallel()
	service := authmock.NewMockUseCase(gomock.NewController(t))
	user := entity.User{ID: "11111111-1111-4111-8111-111111111111", Email: "test@example.test", PasswordHash: "private-password-hash"}
	service.EXPECT().Login(gomock.Any(), user.Email, "correct-password").Return(auth.Result{
		User: user, Access: "private-access-token", Refresh: "private-refresh-token",
		AccessExpires: time.Now().Add(time.Minute), RefreshExpires: time.Now().Add(time.Hour),
	}, nil)

	router := chi.NewRouter()
	NewAuth(service, config.Auth{CookieSecure: true}).RegisterPublic(router)

	recorded := httptest.NewRecorder()
	router.ServeHTTP(recorded, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"test@example.test","password":"correct-password"}`)))

	if recorded.Code != http.StatusOK {
		t.Fatalf("incorrect login response: %d %s", recorded.Code, recorded.Body.String())
	}

	result := recorded.Result()
	defer result.Body.Close()

	cookies := result.Cookies()
	if len(cookies) != 2 {
		t.Fatalf("expected two session cookies, received %d", len(cookies))
	}

	for _, cookie := range cookies {
		if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode {
			t.Fatalf("session cookie flags lost: %+v", cookie)
		}

		switch cookie.Name {
		case "sf_access":
			if cookie.Value != "private-access-token" || cookie.Path != "/api" {
				t.Fatal("incorrect access cookie")
			}
		case "sf_refresh":
			if cookie.Value != "private-refresh-token" || cookie.Path != "/api/v1/auth" {
				t.Fatal("incorrect refresh cookie")
			}
		default:
			t.Fatalf("unexpected session cookie: %s", cookie.Name)
		}
	}

	for _, secret := range []string{"private-access-token", "private-refresh-token", "private-password-hash"} {
		if strings.Contains(recorded.Body.String(), secret) {
			t.Fatal("login exposed a secret in its response body")
		}
	}
}
