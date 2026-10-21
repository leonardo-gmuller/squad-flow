package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	workpackage "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/package"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/response"
)

func RegisterPackages(router chi.Router, svc workpackage.UseCaseInterface) {
	router.Get("/api/v1/packages", func(w http.ResponseWriter, r *http.Request) {
		workPackage, err := svc.List(r.Context())
		if err != nil {
			response.SendContext(r.Context(), w, response.AppError(err))

			return
		}

		response.SendContext(r.Context(), w, response.OK(workPackage))
	})
	router.Get("/api/v1/packages/{id}", func(w http.ResponseWriter, r *http.Request) {
		workPackage, err := svc.Get(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			response.SendContext(r.Context(), w, response.AppError(err))

			return
		}

		response.SendContext(r.Context(), w, response.OK(workPackage))
	})
	router.Post("/api/v1/packages", func(w http.ResponseWriter, r *http.Request) {
		var in workpackage.Input
		if !decode(w, r, &in) {
			return
		}

		workPackage, err := svc.Create(r.Context(), in)
		if err != nil {
			response.SendContext(r.Context(), w, response.AppError(err))

			return
		}

		response.SendContext(r.Context(), w, response.Created(workPackage))
	})
}
