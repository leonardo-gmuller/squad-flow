package middleware

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/telemetry"
)

func Observability(next http.Handler) http.Handler {
	operation := telemetry.NewOperation("http.server")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Health probes and long-lived WebSocket connections do not describe HTTP latency.
		if r.URL.Path == "/healthz" || r.URL.Path == "/api/v1/ws" {
			next.ServeHTTP(w, r)

			return
		}

		method := r.Method
		switch method {
		case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "CONNECT", "TRACE":
		default:
			method = "_OTHER"
		}

		ctx := telemetry.Extract(r.Context(), r.Header.Get("Traceparent"), r.Header.Get("Tracestate"))

		ctx, span := otel.Tracer(telemetry.Scope).Start(ctx, method, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		if span.SpanContext().IsValid() {
			w.Header().Set("X-Trace-Id", span.SpanContext().TraceID().String())
		}

		response := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		defer func() {
			route := "unmatched"
			routeContext := chi.RouteContext(ctx)

			if routeContext != nil {
				route = routeContext.RoutePattern()
			}

			status := response.Status()
			if status == 0 {
				status = http.StatusOK
			}

			if route == "" || route == "/*" {
				route = "unmatched"
			}

			attrs := []attribute.KeyValue{attribute.String("http.request.method", method), attribute.String("http.route", route), attribute.Int("http.response.status_code", status)}
			span.SetName(method + " " + route)
			span.SetAttributes(attrs...)

			outcome := "success"

			if status >= 500 {
				span.SetStatus(codes.Error, "server_error")

				outcome = "error"
			}

			operation.Observe(ctx, start, outcome, attrs...)
		}()
		next.ServeHTTP(response, r.WithContext(ctx))
	})
}
