package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	workpackage "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/package"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/handler/schema"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/response"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
)

type Handler struct{ svc demand.UseCaseInterface }

func actorLabel(user entity.User) string {
	name := user.Name
	if len(name) > 80 {
		name = name[:80]
		for !utf8.ValidString(name) {
			name = name[:len(name)-1]
		}
	}

	return name + " (" + user.ID + ")"
}

func New(svc demand.UseCaseInterface, packageSvc ...workpackage.UseCaseInterface) http.Handler {
	handler := &Handler{svc: svc}

	mux := chi.NewRouter()
	mux.NotFound(response.Handle(func(_ *http.Request) *response.Response { return response.AppError(erring.ErrNotFound) }))
	mux.MethodNotAllowed(response.Handle(func(_ *http.Request) *response.Response { return response.AppError(erring.ErrMethodNotAllowed) }))

	if len(packageSvc) > 0 && packageSvc[0] != nil {
		RegisterPackages(mux, packageSvc[0])
	}

	mux.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		response.Send(w, response.OK(map[string]string{"status": "ok"}))
	})
	mux.Get("/api/v1/demands", handler.list)
	mux.Post("/api/v1/demands", handler.create)
	mux.Put("/api/v1/demands/{id}", handler.edit)
	mux.Get("/api/v1/demands/{id}", handler.detail)
	mux.Post("/api/v1/demands/{id}/move", handler.move)
	mux.Post("/api/v1/demands/{id}/tests", handler.addTest)
	mux.Get("/api/v1/metrics", handler.metrics)
	mux.Get("/api/v1/reports/demands.csv", handler.csv)

	return mux
}

func decode(w http.ResponseWriter, r *http.Request, value any) bool {
	if err := response.Decode(r, value); err != nil {
		response.Send(w, response.AppError(err))

		return false
	}

	return true
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.List(r.Context(), r.URL.Query().Get("kind"), r.URL.Query().Get("status"), r.URL.Query().Get("q"))
	if err != nil {
		response.Send(w, response.AppError(err))

		return
	}

	response.Send(w, response.OK(rows))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req schema.CreateDemandRequest
	if !decode(w, r, &req) {
		return
	}

	out, err := h.svc.Create(r.Context(), req.ToInput())
	if err != nil {
		response.Send(w, response.AppError(err))

		return
	}

	response.Send(w, response.Created(out))
}

func (h *Handler) edit(w http.ResponseWriter, r *http.Request) {
	var in demand.EditInput
	if !decode(w, r, &in) {
		return
	}

	out, err := h.svc.Edit(r.Context(), chi.URLParam(r, "id"), in)
	if err != nil {
		response.Send(w, response.AppError(err))

		return
	}

	response.Send(w, response.OK(out))
}

func (h *Handler) detail(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.Detail(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		response.Send(w, response.AppError(err))

		return
	}

	response.Send(w, response.OK(out))
}

func (h *Handler) move(w http.ResponseWriter, r *http.Request) {
	var req schema.MoveDemandRequest
	if !decode(w, r, &req) {
		return
	}

	input := req.ToInput()
	if user, ok := ctxkey.GetUser(r.Context()); ok {
		input.Actor = actorLabel(user)
		input.ActorID = user.ID
	}

	out, err := h.svc.Move(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		response.Send(w, response.AppError(err))

		return
	}

	response.Send(w, response.OK(out))
}

func (h *Handler) addTest(w http.ResponseWriter, r *http.Request) {
	var req schema.TestCaseRequest
	if !decode(w, r, &req) {
		return
	}

	input := req.ToInput()
	if user, ok := ctxkey.GetUser(r.Context()); ok {
		input.Author = actorLabel(user)
		input.AuthorID = user.ID
	}

	out, err := h.svc.AddTest(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		response.Send(w, response.AppError(err))

		return
	}

	response.Send(w, response.Created(out))
}

func (h *Handler) metrics(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.Metrics(r.Context())
	if err != nil {
		response.Send(w, response.AppError(err))

		return
	}

	response.Send(w, response.OK(out))
}

func spreadsheetSafe(session string) string {
	trimmed := strings.TrimLeft(session, " \t\r\n")
	if strings.HasPrefix(trimmed, "=") || strings.HasPrefix(trimmed, "+") || strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "@") {
		return "'" + session
	}

	return session
}

func (h *Handler) csv(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.List(r.Context(), "", "", "")
	if err != nil {
		response.Send(w, response.AppError(err))

		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="squadflow-demandas.csv"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte{0xef, 0xbb, 0xbf}) // Excel UTF-8 BOM
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"ID", "Título", "Esteira", "Status", "Prioridade", "Responsável", "QA", "Referência", "OS agrupadora", "Produto", "Critérios de aceite", "Prazo", "Eventos", "Devoluções", "Criado em", "Concluído em", "Versão"})

	for _, item := range rows {
		completed := ""
		if item.CompletedAt != nil {
			completed = item.CompletedAt.Format(time.RFC3339)
		}

		_ = writer.Write([]string{fmt.Sprintf("SF-%04d", item.Number), spreadsheetSafe(item.Title), spreadsheetSafe(item.KindName), string(item.Status), string(item.Priority), spreadsheetSafe(item.Assignee), spreadsheetSafe(item.Reviewer), spreadsheetSafe(item.Reference), spreadsheetSafe(item.PackageID), spreadsheetSafe(item.Product), spreadsheetSafe(item.AcceptanceCriteria), func() string {
			if item.DueDate != nil {
				return item.DueDate.Format("2006-01-02")
			}

			return ""
		}(), spreadsheetSafe(strings.Join(item.EventCodes, ", ")), strconv.Itoa(item.ReturnCount), item.CreatedAt.Format(time.RFC3339), completed, spreadsheetSafe(item.ReleaseName)})
	}

	writer.Flush()
}
