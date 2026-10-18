package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/middleware"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/response"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
)

func (h *AuthHandler) registerPasswordReset(mux chi.Router) {
	limiter := middleware.NewRateLimiter(30, 5*time.Minute)
	mux.With(limiter.Limit).Post("/api/v1/auth/password-reset", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Token    string `json:"token"`
			Password string `json:"password"`
		}

		if !decode(w, r, &input) {
			return
		}

		userID, err := h.Service.ResetPassword(r.Context(), input.Token, input.Password)
		if err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		response.Send(w, response.OK(map[string]any{"ok": true, "user_id": userID}))
	})
}

func (h *AuthHandler) registerPasswordManagement(mux chi.Router) {
	mux.Put("/api/v1/auth/password", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Current  string `json:"current_password"`
			Password string `json:"password"`
		}

		if !decode(w, r, &input) {
			return
		}

		actor, _ := ctxkey.GetUser(r.Context())
		if err := h.Service.ChangePassword(r.Context(), actor, input.Current, input.Password); err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		h.clear(w)
		response.Send(w, response.OK(map[string]bool{"ok": true}))
	})
	mux.Put("/api/v1/users/{id}/password", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Password string `json:"password"`
		}

		if !decode(w, r, &input) {
			return
		}

		actor, _ := ctxkey.GetUser(r.Context())
		userID := chi.URLParam(r, "id")

		if err := h.Service.AdminChangePassword(r.Context(), actor, userID, input.Password); err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		if userID == actor.ID {
			h.clear(w)
		}

		response.Send(w, response.OK(map[string]bool{"ok": true}))
	})
	mux.Post("/api/v1/users/{id}/password-reset", func(w http.ResponseWriter, r *http.Request) {
		actor, _ := ctxkey.GetUser(r.Context())

		out, err := h.Service.CreatePasswordReset(r.Context(), actor, chi.URLParam(r, "id"))
		if err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		response.Send(w, response.Created(out))
	})
}
