package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/response"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/openai"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/reporthtml"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

type ReportHandler struct {
	Service report.UseCaseInterface
	Jobs    report.JobUseCaseInterface
	Enabled bool
}

func (h *ReportHandler) Register(mux chi.Router) {
	mux.Get("/api/v1/reports/status", func(w http.ResponseWriter, _ *http.Request) {
		response.Send(w, response.OK(map[string]any{"enabled": h.Enabled, "asynchronous": true, "format": "html"}))
	})
	mux.Get("/api/v1/reports/jobs", func(w http.ResponseWriter, r *http.Request) {
		user, _ := ctxkey.GetUser(r.Context())

		jobs, err := h.Jobs.List(r.Context(), user.ID)
		if err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		response.Send(w, response.OK(jobs))
	})
	mux.Get("/api/v1/reports/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		user, _ := ctxkey.GetUser(r.Context())

		job, err := h.Jobs.Get(r.Context(), user.ID, chi.URLParam(r, "id"))
		if err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		response.Send(w, response.OK(job))
	})
	mux.Post("/api/v1/reports/jobs/{id}/retry", func(w http.ResponseWriter, r *http.Request) {
		if !h.Enabled {
			response.Send(w, response.AppError(erring.ErrServiceUnavailable.WithMessage(openai.ErrUnavailable.Error())))

			return
		}

		user, _ := ctxkey.GetUser(r.Context())

		job, err := h.Jobs.Retry(r.Context(), user, chi.URLParam(r, "id"), r.Header.Get("Idempotency-Key"))
		if err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		w.Header().Set("Location", "/api/v1/reports/jobs/"+job.ID)
		response.Send(w, response.Accepted(job))
	})
	mux.Get("/api/v1/reports", func(w http.ResponseWriter, r *http.Request) {
		user, _ := ctxkey.GetUser(r.Context())

		id := r.URL.Query().Get("demand_id")
		if id != "" && !identifier.IsUUID(id) {
			response.Send(w, response.AppError(erring.ErrInvalidInput.WithMessage("demanda inválida")))

			return
		}

		out, err := h.Service.List(r.Context(), user.ID, id)
		if err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		response.Send(w, response.OK(out))
	})
	mux.Get("/api/v1/reports/{id}", func(w http.ResponseWriter, r *http.Request) {
		user, _ := ctxkey.GetUser(r.Context())

		id := chi.URLParam(r, "id")
		if !identifier.IsUUID(id) {
			response.Send(w, response.AppError(erring.ErrInvalidInput.WithMessage("relatório inválido")))

			return
		}

		out, err := h.Service.Get(r.Context(), user.ID, id)
		if err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		response.Send(w, response.OK(out))
	})
	mux.Get("/api/v1/reports/{id}/html", h.document)
	mux.Post("/api/v1/reports/generate", func(w http.ResponseWriter, r *http.Request) {
		if !h.Enabled {
			response.Send(w, response.AppError(erring.ErrServiceUnavailable.WithMessage(openai.ErrUnavailable.Error())))

			return
		}

		var input report.Request
		if !decode(w, r, &input) {
			return
		}

		user, _ := ctxkey.GetUser(r.Context())

		job, err := h.Jobs.Submit(r.Context(), user, input, r.Header.Get("Idempotency-Key"))
		if err != nil {
			response.Send(w, response.AppError(err))

			return
		}

		w.Header().Set("Location", "/api/v1/reports/jobs/"+job.ID)
		response.Send(w, response.Accepted(job))
	})
}

func (h *ReportHandler) document(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !identifier.IsUUID(id) {
		response.Send(w, response.AppError(erring.ErrInvalidInput.WithMessage("relatório inválido")))

		return
	}

	user, _ := ctxkey.GetUser(r.Context())

	item, err := h.Service.Get(r.Context(), user.ID, id)
	if err != nil {
		response.Send(w, response.AppError(err))

		return
	}

	document, err := (reporthtml.Renderer{}).Render(item.Content)
	if err != nil {
		response.Send(w, response.AppError(err))

		return
	}

	disposition := "inline"
	if r.URL.Query().Get("download") == "1" {
		disposition = "attachment"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=\"relatorio-%s.html\"", disposition, id))
	w.Header().Set("Content-Security-Policy", reporthtml.ContentSecurityPolicy+"; sandbox; frame-ancestors 'self'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = w.Write([]byte(document))
}
