package handler

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/config"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/auth"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/response"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
)

type AuthHandler struct {
	Service  auth.UseCaseInterface
	Config   config.Auth
	mu       sync.Mutex
	attempts map[string]attempt
}
type attempt struct {
	count   int
	expires time.Time
}

func NewAuth(service auth.UseCaseInterface, cfg config.Auth) *AuthHandler {
	return &AuthHandler{Service: service, Config: cfg, attempts: map[string]attempt{}}
}

func (h *AuthHandler) sessionCookie(name, value, path string, expires time.Time) *http.Cookie {
	age := int(time.Until(expires).Seconds())
	if value == "" {
		age = -1
	}

	return &http.Cookie{Name: name, Value: value, Path: path, Expires: expires, MaxAge: age, HttpOnly: true, Secure: h.Config.CookieSecure, SameSite: http.SameSiteStrictMode}
}

func (h *AuthHandler) cookie(w http.ResponseWriter, name, value, path string, expires time.Time) {
	http.SetCookie(w, h.sessionCookie(name, value, path, expires))
}

func (h *AuthHandler) sessionResult(result auth.Result) *response.Response {
	return response.OK(map[string]any{"user": result.User, "access_expires_at": result.AccessExpires}).WithCookies(
		h.sessionCookie("sf_access", result.Access, "/api", result.AccessExpires),
		h.sessionCookie("sf_refresh", result.Refresh, "/api/v1/auth", result.RefreshExpires),
	)
}

func (h *AuthHandler) result(w http.ResponseWriter, result auth.Result) {
	response.Send(w, h.sessionResult(result))
}

func refreshCookie(r *http.Request) string {
	c, err := r.Cookie("sf_refresh")
	if err != nil {
		return ""
	}

	return c.Value
}

func (h *AuthHandler) RegisterPublic(mux chi.Router) {
	h.registerPasswordReset(mux)
	mux.Post("/api/v1/auth/login", response.Handle(h.login))
	mux.Post("/api/v1/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		result, err := h.Service.Refresh(r.Context(), refreshCookie(r))
		if err != nil {
			h.clear(w)
			response.Send(w, response.AppError(err))

			return
		}

		h.result(w, result)
	})
	mux.Post("/api/v1/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		err := h.Service.Logout(r.Context(), refreshCookie(r))
		h.clear(w)

		if err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		response.Send(w, response.OK(map[string]bool{"ok": true}))
	})
}

func (h *AuthHandler) RegisterProtected(mux chi.Router) {
	h.registerPasswordManagement(mux)
	mux.Get("/api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		u, _ := ctxkey.GetUser(r.Context())
		response.Send(w, response.OK(map[string]any{"user": u}))
	})
	mux.Get("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		u, _ := ctxkey.GetUser(r.Context())

		out, err := h.Service.Users(r.Context(), u)
		if err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		response.Send(w, response.OK(out))
	})
	mux.Post("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if !decode(w, r, &input) {
			return
		}

		u, _ := ctxkey.GetUser(r.Context())

		out, err := h.Service.CreateUser(r.Context(), u, input.Name, input.Email, input.Password)
		if err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		response.Send(w, response.Created(out))
	})
}

func (h *AuthHandler) clear(w http.ResponseWriter) {
	h.cookie(w, "sf_access", "", "/api", time.Unix(1, 0))
	h.cookie(w, "sf_refresh", "", "/api/v1/auth", time.Unix(1, 0))
}

func (h *AuthHandler) login(r *http.Request) *response.Response {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := response.Decode(r, &input); err != nil {
		return response.AppError(err)
	}

	if len(input.Email) > 254 || len(input.Password) > 72 {
		return response.AppError(erring.ErrInvalidCredentials)
	}

	if !h.allowLogin(r, input.Email) {
		return response.AppError(erring.ErrTooManyRequests).WithHeaders(map[string]string{"Retry-After": "300"})
	}

	result, err := h.Service.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		return response.AppError(err)
	}

	return h.sessionResult(result)
}

func (h *AuthHandler) allowLogin(r *http.Request, email string) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	for key, value := range h.attempts {
		if now.After(value.expires) {
			delete(h.attempts, key)
		}
	}

	if len(h.attempts) >= 10000 {
		return false
	}

	allowed := true
	// Keep separate account and source-address limits behind shared proxies.
	for _, limit := range []struct {
		key string
		max int
	}{{host, 150}, {host + ":" + strings.ToLower(strings.TrimSpace(email)), 15}} {
		attempt := h.attempts[limit.key]
		if attempt.count == 0 {
			attempt.expires = now.Add(5 * time.Minute)
		}

		attempt.count++
		h.attempts[limit.key] = attempt

		if attempt.count > limit.max {
			allowed = false
		}
	}

	return allowed
}
