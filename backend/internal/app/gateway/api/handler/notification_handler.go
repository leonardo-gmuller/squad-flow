package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/notification"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/middleware"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/response"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
)

func RegisterNotifications(mux chi.Router, service notification.UseCaseInterface, changed func(string)) {
	mux.Route("/api/v1/notifications", func(routes chi.Router) {
		routes.Use(middleware.RequireMember)
		routes.Get("/", func(w http.ResponseWriter, r *http.Request) {
			user, _ := ctxkey.GetUser(r.Context())
			query := r.URL.Query()
			limit := 20

			if query.Has("limit") {
				value, err := strconv.Atoi(query.Get("limit"))
				if err != nil || value <= 0 {
					response.SendContext(r.Context(), w, response.AppError(erring.ErrInvalidInput.WithMessage("limite inválido")))

					return
				}

				limit = value
			}

			page, err := service.List(r.Context(), user, query.Get("filter"), limit, query.Get("before"))
			if err != nil {
				response.SendContext(r.Context(), w, response.AppError(err))

				return
			}

			response.SendContext(r.Context(), w, response.OK(page))
		})
		routes.Get("/unread-count", func(w http.ResponseWriter, r *http.Request) {
			user, _ := ctxkey.GetUser(r.Context())

			count, err := service.UnreadCount(r.Context(), user)
			if err != nil {
				response.SendContext(r.Context(), w, response.AppError(err))

				return
			}

			response.SendContext(r.Context(), w, response.OK(map[string]int{"count": count}))
		})
		routes.Post("/{id}/read", func(w http.ResponseWriter, r *http.Request) {
			user, _ := ctxkey.GetUser(r.Context())

			readAt, err := service.Read(r.Context(), user, chi.URLParam(r, "id"))
			if err != nil {
				response.SendContext(r.Context(), w, response.AppError(err))

				return
			}

			if changed != nil {
				changed(user.ID)
			}

			response.SendContext(r.Context(), w, response.OK(map[string]any{"read_at": readAt}))
		})
		routes.Post("/read-all", func(w http.ResponseWriter, r *http.Request) {
			user, _ := ctxkey.GetUser(r.Context())

			readAt, err := service.ReadAll(r.Context(), user)
			if err != nil {
				response.SendContext(r.Context(), w, response.AppError(err))

				return
			}

			if changed != nil {
				changed(user.ID)
			}

			response.SendContext(r.Context(), w, response.OK(map[string]any{"read_at": readAt}))
		})
	})
}
