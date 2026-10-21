package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/middleware"
)

//nolint:paralleltest // Installs a process-wide tracing provider for HTTP instrumentation.
func TestObservabilityUsesRouteTemplatesPropagatesTraceAndRecordsRecoveredPanic(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()), sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(provider)
	t.Cleanup(func() { _ = provider.Shutdown(t.Context()); otel.SetTracerProvider(tracenoop.NewTracerProvider()) })

	router := chi.NewRouter()
	router.Use(middleware.Observability, chimiddleware.Recoverer)
	router.Get("/items/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	router.Get("/panic", func(_ http.ResponseWriter, _ *http.Request) { panic("private panic") })

	parent := "00-11111111111111111111111111111111-2222222222222222-01"
	request := httptest.NewRequest(http.MethodGet, "/items/private-id?token=private-token", nil)
	request.Header.Set("Traceparent", parent)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent || response.Header().Get("X-Trace-Id") != "11111111111111111111111111111111" {
		t.Fatal("trace context was not propagated")
	}

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/unmatched/private-id", nil))
	panicResponse := httptest.NewRecorder()
	router.ServeHTTP(panicResponse, httptest.NewRequest(http.MethodGet, "/panic", nil))

	spans := recorder.Ended()
	if len(spans) != 3 || spans[0].Name() != "GET /items/{id}" || spans[1].Name() != "GET unmatched" || panicResponse.Code != 500 {
		t.Fatalf("incorrect HTTP traces: %+v", spans)
	}

	if spans[0].Parent().SpanID().String() != "2222222222222222" {
		t.Fatal("remote parent lost")
	}

	for _, span := range spans {
		for _, attr := range span.Attributes() {
			if strings.Contains(attr.Value.Emit(), "private") {
				t.Fatal("raw URL, query or path identifier leaked to telemetry")
			}
		}
	}

	if spans[2].Status().Description != "server_error" {
		t.Fatal("recovered panic not marked as a server error")
	}
}
