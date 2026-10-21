package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer logRequest(r.Context(), r, wrapped, start)
		next.ServeHTTP(wrapped, r)
	})
}

func logRequest(ctx context.Context, request *http.Request, response chimiddleware.WrapResponseWriter, start time.Time) {
	id, _ := ctxkey.GetRequestID(ctx)

	status := response.Status()
	if status == 0 {
		status = http.StatusOK
	}

	slog.InfoContext(ctx, "http request", "request_id", id, "method", request.Method, "path", request.URL.Path, "status", status, "duration", time.Since(start).String())
}
