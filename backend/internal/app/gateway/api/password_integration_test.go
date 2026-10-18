//go:build integration

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/auth"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/security"
)

func assertPasswordEndpoints(t *testing.T, admin, member *http.Client, baseURL string, db *postgres.Client, user entity.User) {
	t.Helper()

	call := func(client *http.Client, method, path string, body any, status int) []byte {
		return callAPI(t, client, baseURL, method, path, body, status)
	}
	login := func(password string, status int) {
		call(member, "POST", "/api/v1/auth/login", map[string]string{"email": user.Email, "password": password}, status)
	}
	target := "/api/v1/users/" + user.ID
	reset := func() auth.PasswordReset {
		raw := call(admin, "POST", target+"/password-reset", map[string]any{}, 201)

		var result auth.PasswordReset
		if err := json.Unmarshal(raw, &result); err != nil {
			t.Fatal(err)
		}

		if len(result.Token) != 43 || time.Until(result.ExpiresAt) < 29*time.Minute || time.Until(result.ExpiresAt) > 31*time.Minute {
			t.Fatal("invalid reset token or expiration")
		}

		return result
	}
	consume := func(token, password string, status int) {
		call(testClient(t), "POST", "/api/v1/auth/password-reset", map[string]string{"token": token, "password": password}, status)
	}

	call(testClient(t), "PUT", target+"/password", map[string]string{"password": "AdminAlterada123!"}, 401)
	call(member, "PUT", target+"/password", map[string]string{"password": "AdminAlterada123!"}, 403)
	call(member, "POST", target+"/password-reset", map[string]any{}, 403)
	call(admin, "PUT", "/api/v1/users/not-a-uuid/password", map[string]string{"password": "AdminAlterada123!"}, 400)
	call(admin, "PUT", target+"/password", map[string]string{"password": "short"}, 400)
	call(member, "GET", "/api/v1/auth/me", nil, 200)

	first := reset()
	second := reset()

	consume(first.Token, "ResetEscolhida123!", 400)
	call(member, "GET", "/api/v1/auth/me", nil, 200)

	hash := (security.Tokens{}).HashRefresh("password-reset:" + second.Token)

	var stored string

	if err := db.Pool.QueryRow(t.Context(), `SELECT token_hash FROM password_reset_tokens WHERE token_hash=$1`, hash).Scan(&stored); err != nil || stored == second.Token {
		t.Fatal("reset token was not stored as a hash")
	}

	usersJSON := call(admin, "GET", "/api/v1/users", nil, 200)
	if bytes.Contains(usersJSON, []byte(second.Token)) || bytes.Contains(usersJSON, []byte("password_hash")) {
		t.Fatal("user listing exposes credential data")
	}

	apiURL, _ := url.Parse(baseURL + "/api/v1/auth")
	oldRefresh := refreshToken(t, member, apiURL)

	call(admin, "PUT", target+"/password", map[string]string{"password": "AdminAlterada123!"}, 200)
	call(member, "GET", "/api/v1/auth/me", nil, 401)
	assertRefreshRevoked(t, baseURL, oldRefresh)
	consume(second.Token, "ResetEscolhida123!", 400)
	login("member-test-password", 401)
	login("AdminAlterada123!", 200)

	expired := reset()

	expiredHash := (security.Tokens{}).HashRefresh("password-reset:" + expired.Token)
	if _, err := db.Pool.Exec(t.Context(), `UPDATE password_reset_tokens SET created_at=now()-interval '2 minutes',expires_at=now()-interval '1 minute' WHERE token_hash=$1`, expiredHash); err != nil {
		t.Fatal(err)
	}

	consume(expired.Token, "ResetEscolhida123!", 400)
	consume(strings.Repeat("x", 43), "ResetEscolhida123!", 400)

	current := reset()
	consume(current.Token, "short", 400)

	oldRefresh = refreshToken(t, member, apiURL)

	consume(current.Token, "ResetEscolhida123!", 200)
	consume(current.Token, "NaoPodeReusar123!", 400)
	call(member, "GET", "/api/v1/auth/me", nil, 401)
	assertRefreshRevoked(t, baseURL, oldRefresh)
	login("AdminAlterada123!", 401)
	login("ResetEscolhida123!", 200)

	call(member, "PUT", "/api/v1/auth/password", map[string]string{"current_password": "wrong", "password": "MinhaNovaSenha123!"}, 400)
	call(member, "GET", "/api/v1/auth/me", nil, 200)
	oldRefresh = refreshToken(t, member, apiURL)
	call(member, "PUT", "/api/v1/auth/password", map[string]string{"current_password": "ResetEscolhida123!", "password": "MinhaNovaSenha123!"}, 200)
	call(member, "GET", "/api/v1/auth/me", nil, 401)
	assertRefreshRevoked(t, baseURL, oldRefresh)
	login("ResetEscolhida123!", 401)
	login("MinhaNovaSenha123!", 200)

	var auditCount int
	if err := db.Pool.QueryRow(t.Context(), `SELECT count(*) FROM user_password_events WHERE user_id=$1`, user.ID).Scan(&auditCount); err != nil || auditCount < 7 {
		t.Fatal("credential audit is missing")
	}

	call(admin, "PUT", target+"/password", map[string]string{"password": "member-test-password"}, 200)
	login("member-test-password", 200)
}

func assertRefreshRevoked(t *testing.T, baseURL, raw string) {
	t.Helper()

	client := testClient(t)
	apiURL, _ := url.Parse(baseURL + "/api/v1/auth")
	client.Jar.SetCookies(apiURL, []*http.Cookie{{Name: "sf_refresh", Value: raw, Path: "/api/v1/auth"}})
	callAPI(t, client, baseURL, "POST", "/api/v1/auth/refresh", map[string]any{}, 401)
}
