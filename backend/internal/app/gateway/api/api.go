package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/config"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/handler"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/middleware"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/response"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/openai"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/realtime"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
)

type API struct {
	Handler http.Handler
	Hub     *realtime.Hub
}

func New(application *app.App, cfg config.Config) *API {
	mux := chi.NewRouter()
	mux.NotFound(response.Handle(func(_ *http.Request) *response.Response { return response.AppError(erring.ErrNotFound) }))
	mux.MethodNotAllowed(response.Handle(func(_ *http.Request) *response.Response { return response.AppError(erring.ErrMethodNotAllowed) }))

	authHTTP := handler.NewAuth(application.Auth, cfg.Auth)
	hub := realtime.New()
	origin := middleware.NewOriginPolicy(cfg.Auth.AllowedOrigins)
	domainHTTP := handler.NewWorkspace(application.NewDemandUseCase(), application.NewPackageUseCase(), application.NewReleaseUseCase(), application.NewBoardUseCase(), application.NewCatalogUseCase())
	reportsHTTP := &handler.ReportHandler{Service: application.Reports, Jobs: application.ReportJobs, Enabled: cfg.OpenAI.APIKey != ""}

	mux.Use(chimiddleware.RequestID)
	mux.Use(middleware.RequestContext)
	mux.Use(middleware.Observability)
	mux.Use(middleware.RequestLogger)
	mux.Use(chimiddleware.Recoverer)
	mux.Use(chimiddleware.NoCache)
	mux.Use(origin.Protect)
	mux.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			slog.Warn("health response failed", "error", err)
		}
	})
	authHTTP.RegisterPublic(mux)
	mux.Group(func(secured chi.Router) {
		secured.Use(middleware.RequireAuth(application.Auth))
		authHTTP.RegisterProtected(secured)
		handler.RegisterNotifications(secured, application.NewNotificationUseCase(), func(userID string) { hub.Publish(userID, realtime.Event{Type: "notifications.changed"}) })
		secured.Get("/api/v1/reports/demands.csv", domainHTTP.ServeHTTP)
		reportsHTTP.Register(secured)
		secured.Handle("/api/v1/mcp", handler.MCP{Data: application.NewDemandUseCase()})
		secured.Get("/api/v1/ws", func(w http.ResponseWriter, r *http.Request) {
			user, _ := ctxkey.GetUser(r.Context())
			raw := middleware.AccessToken(r)
			hub.Serve(w, r, user, origin.Allowed, func(ctx context.Context) bool {
				_, err := application.Auth.Authenticate(ctx, raw)

				return err == nil
			}, func(ctx context.Context, user entity.User, in report.Request, key string) (report.Job, error) {
				if !reportsHTTP.Enabled {
					return report.Job{}, openai.ErrUnavailable
				}

				return application.ReportJobs.Submit(ctx, user, in, key)
			})
		})
		secured.Handle("/*", middleware.WorkspaceUpdates(hub)(domainHTTP))
	})

	return &API{Handler: mux, Hub: hub}
}
func (a *API) Close() { a.Hub.Close() }
