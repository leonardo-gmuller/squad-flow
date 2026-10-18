package middleware

import (
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
)

// RequestContext runs after Chi's RequestID and bridges HTTP metadata to ctxkey.
// Idempotency-Key is propagated as metadata; it does not deduplicate mutations.
func RequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := ctxkey.PutRequestID(r.Context(), chimiddleware.GetReqID(r.Context()))
		ctx = ctxkey.PutTransport(ctx, "http")

		if value := r.Header.Get("Authorization"); value != "" {
			ctx = ctxkey.PutAuthorizationHeader(ctx, value)
		}

		if value := r.Header.Get("Idempotency-Key"); value != "" {
			ctx = ctxkey.PutIdempotencyKey(ctx, value)
		}

		id, _ := ctxkey.GetRequestID(ctx)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
