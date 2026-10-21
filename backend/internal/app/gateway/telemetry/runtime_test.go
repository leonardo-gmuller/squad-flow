package telemetry

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	logpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/config"
)

//nolint:paralleltest // Installs process-wide OpenTelemetry providers and slog handler.
func TestRuntimeExportsTracesMetricsAndCorrelatedLogs(t *testing.T) {
	var lock sync.Mutex

	paths := map[string]bool{}

	var logs logpb.ExportLogsServiceRequest

	var traces tracepb.ExportTraceServiceRequest

	var metrics metricpb.ExportMetricsServiceRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		lock.Lock()
		defer lock.Unlock()

		paths[r.URL.Path] = true

		switch r.URL.Path {
		case "/v1/traces":
			err = proto.Unmarshal(payload, &traces)
		case "/v1/metrics":
			err = proto.Unmarshal(payload, &metrics)
		case "/v1/logs":
			err = proto.Unmarshal(payload, &logs)
		default:
			t.Errorf("unexpected OTLP endpoint: %s", r.URL.Path)
		}

		if err != nil {
			t.Error(err)
		}

		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	defer server.Close()

	previous := slog.Default()

	t.Cleanup(func() {
		slog.SetDefault(previous)
		otel.SetTracerProvider(tracenoop.NewTracerProvider())
		otel.SetMeterProvider(metricnoop.NewMeterProvider())
	})

	runtime, err := Setup(t.Context(), config.Config{Environment: config.EnvTest, App: config.App{Version: "test"}, Reports: config.Reports{Role: "api"}, Telemetry: config.Telemetry{Enabled: true, Endpoint: server.URL, SampleRatio: 1, MetricInterval: time.Hour, LogLevel: "info"}})
	if err != nil {
		t.Fatal(err)
	}

	ctx, span := otel.Tracer(Scope).Start(t.Context(), "export-verification")
	spanContext := span.SpanContext()
	traceID := spanContext.TraceID()
	spanID := spanContext.SpanID()

	NewOperation("verification").Observe(ctx, time.Now(), "success")
	slog.InfoContext(ctx, "correlated OTLP log")
	span.End()

	shutdownCtx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	if err = runtime.Shutdown(shutdownCtx); err != nil {
		t.Fatal(err)
	}

	lock.Lock()
	defer lock.Unlock()

	for _, path := range []string{"/v1/traces", "/v1/metrics", "/v1/logs"} {
		if !paths[path] {
			t.Errorf("signal was not exported: %s", path)
		}
	}

	if len(traces.GetResourceSpans()) == 0 || len(metrics.GetResourceMetrics()) == 0 || len(logs.GetResourceLogs()) == 0 {
		t.Fatal("empty OTLP signal")
	}

	found := false

	for _, resource := range logs.GetResourceLogs() {
		for _, scope := range resource.GetScopeLogs() {
			for _, record := range scope.GetLogRecords() {
				if record.GetBody().GetStringValue() == "correlated OTLP log" {
					found = bytes.Equal(record.GetTraceId(), traceID[:]) && bytes.Equal(record.GetSpanId(), spanID[:])
				}
			}
		}
	}

	if !found {
		t.Fatal("OTLP log lost native trace/span correlation")
	}
}
