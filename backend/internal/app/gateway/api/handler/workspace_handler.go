package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/catalog"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	workpackage "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/package"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/release"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/response"
)

func NewWorkspace(demands demand.UseCaseInterface, packages workpackage.UseCaseInterface, releases release.UseCaseInterface, board demand.BoardUseCaseInterface, catalogs ...catalog.UseCaseInterface) http.Handler {
	mux := chi.NewRouter()
	if len(catalogs) > 0 {
		RegisterCatalog(mux, catalogs[0])
	}

	mux.Get("/api/v1/releases", func(w http.ResponseWriter, r *http.Request) {
		items, err := releases.List(r.Context())
		if err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		response.Send(w, response.OK(items))
	})
	mux.Post("/api/v1/releases", func(w http.ResponseWriter, r *http.Request) {
		var input release.Input
		if !decode(w, r, &input) {
			return
		}

		item, err := releases.Create(r.Context(), input)
		if err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		response.Send(w, response.Created(item))
	})
	mux.Put("/api/v1/releases/{id}", func(w http.ResponseWriter, r *http.Request) {
		var input release.Input
		if !decode(w, r, &input) {
			return
		}

		if err := releases.Update(r.Context(), chi.URLParam(r, "id"), input); err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		response.Send(w, response.OK(map[string]bool{"ok": true}))
	})
	mux.Post("/api/v1/releases/{id}/archive", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Archived *bool `json:"archived"`
		}

		if !decode(w, r, &input) {
			return
		}

		if input.Archived == nil {
			response.Send(w, response.AppError(erring.ErrInvalidInput.WithMessage("informe se a versão deve ser arquivada")))

			return
		}

		if err := releases.SetArchived(r.Context(), chi.URLParam(r, "id"), *input.Archived); err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		response.Send(w, response.OK(map[string]bool{"ok": true}))
	})
	mux.Get("/api/v1/board", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		limit, offset := 20, 0

		var err error

		if query.Has("limit") {
			limit, err = strconv.Atoi(query.Get("limit"))
			if err != nil || limit == 0 {
				response.Send(w, response.AppError(erring.ErrInvalidInput.WithMessage("limite inválido")))

				return
			}
		}

		if query.Has("offset") {
			offset, err = strconv.Atoi(query.Get("offset"))
			if err != nil {
				response.Send(w, response.AppError(erring.ErrInvalidInput.WithMessage("página inválida")))

				return
			}
		}

		out, err := board.Board(r.Context(), demand.BoardFilter{ReleaseID: query.Get("release_id"), Archive: query.Get("archive"), Kind: query.Get("kind"), Owner: query.Get("owner"), Search: query.Get("q"), Status: query.Get("status"), Limit: limit, Offset: offset})
		if err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		response.Send(w, response.OK(out))
	})
	mux.Mount("/", New(demands, packages))

	return mux
}
