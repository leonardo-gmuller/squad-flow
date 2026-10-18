package middleware

import (
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/realtime"
)

type Publisher interface {
	Publish(userID string, event realtime.Event)
}

func WorkspaceUpdates(publisher Publisher) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorded := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(recorded, r)

			status := recorded.Status()
			if status == 0 {
				status = http.StatusOK
			}

			if (r.Method == http.MethodPost || r.Method == http.MethodPut) && status >= 200 && status < 300 {
				publisher.Publish("", realtime.Event{Type: "workspace.updated", Data: map[string]string{"path": r.URL.Path}})
			}
		})
	}
}
