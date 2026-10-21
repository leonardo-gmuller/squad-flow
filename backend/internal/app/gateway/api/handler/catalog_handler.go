package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/auth"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/catalog"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/response"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
)

func RegisterCatalog(mux chi.Router, service catalog.UseCaseInterface) {
	mux.Get("/api/v1/workstreams", func(w http.ResponseWriter, r *http.Request) {
		items, err := service.List(r.Context())
		if err != nil {
			response.SendContext(r.Context(), w, response.AppError(err))

			return
		}

		response.SendContext(r.Context(), w, response.OK(items))
	})
	mux.Get("/api/v1/members", func(w http.ResponseWriter, r *http.Request) {
		items, err := service.Members(r.Context())
		if err != nil {
			response.SendContext(r.Context(), w, response.AppError(err))

			return
		}

		response.SendContext(r.Context(), w, response.OK(items))
	})

	save := func(create bool) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user, ok := ctxkey.GetUser(r.Context())
			if !ok || user.Role != "admin" {
				response.SendContext(r.Context(), w, response.AppError(auth.ErrForbidden))

				return
			}

			var input catalog.Input
			if !decode(w, r, &input) {
				return
			}

			if create {
				item, err := service.Create(r.Context(), user, input)
				if err != nil {
					response.SendContext(r.Context(), w, response.AppError(err))

					return
				}

				response.SendContext(r.Context(), w, response.Created(item))
			} else {
				item, err := service.Update(r.Context(), user, chi.URLParam(r, "id"), input)
				if err != nil {
					response.SendContext(r.Context(), w, response.AppError(err))

					return
				}

				response.SendContext(r.Context(), w, response.OK(item))
			}
		}
	}
	mux.Post("/api/v1/workstreams", save(true))
	mux.Put("/api/v1/workstreams/{id}", save(false))
}
