// Package telemetry configures OpenTelemetry and instruments infrastructure boundaries.
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/config"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

const Scope = "github.com/leonardo-gmuller/squad-flow/backend"

type Runtime struct {
	traces   *sdktrace.TracerProvider
	metrics  *sdkmetric.MeterProvider
	logs     *sdklog.LoggerProvider
	previous *slog.Logger
}

func Setup(ctx context.Context, cfg config.Config) (*Runtime, error) {
	service := cfg.Telemetry.ServiceName
	if service == "" {
		service = "squadflow-" + cfg.Reports.Role
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Telemetry.LogLevel)); err != nil {
		return nil, fmt.Errorf("log level: %w", err)
	}

	local := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}).WithAttrs([]slog.Attr{
		slog.String("service", service), slog.String("environment", string(cfg.Environment)), slog.String("version", cfg.App.Version),
	})
	result := &Runtime{previous: slog.Default()}
	slog.SetDefault(slog.New(&LogHandler{Local: local, Level: level}))

	if !cfg.Telemetry.Enabled {
		return result, nil
	}

	instance, err := identifier.NewUUID()
	if err != nil {
		return nil, errors.Join(erring.Wrap(err, "telemetry.instance"), result.Shutdown(ctx))
	}

	res := resource.NewSchemaless(attribute.String("service.instance.id", instance), attribute.String("service.name", service), attribute.String("service.namespace", "squadflow"), attribute.String("service.version", cfg.App.Version), attribute.String("deployment.environment.name", string(cfg.Environment)))
	// Export errors use only the local sink, avoiding recursive OTLP log exports.
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(_ error) { slog.New(local).Warn("telemetry export failed", "component", "otel-exporter") }))

	traceExporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(strings.TrimRight(cfg.Telemetry.Endpoint, "/")+"/v1/traces"), otlptracehttp.WithTimeout(5*time.Second))
	if err != nil {
		return nil, errors.Join(fmt.Errorf("trace exporter: %w", err), result.Shutdown(ctx))
	}

	result.traces = sdktrace.NewTracerProvider(sdktrace.WithResource(res), sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.Telemetry.SampleRatio))), sdktrace.WithBatcher(traceExporter))
	metricExporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(strings.TrimRight(cfg.Telemetry.Endpoint, "/")+"/v1/metrics"), otlpmetrichttp.WithTimeout(5*time.Second))
	if err != nil {
		return nil, errors.Join(fmt.Errorf("metric exporter: %w", err), result.Shutdown(ctx))
	}

	result.metrics = sdkmetric.NewMeterProvider(sdkmetric.WithResource(res), sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(cfg.Telemetry.MetricInterval))))
	logExporter, err := otlploghttp.New(ctx, otlploghttp.WithEndpointURL(strings.TrimRight(cfg.Telemetry.Endpoint, "/")+"/v1/logs"), otlploghttp.WithTimeout(5*time.Second))
	if err != nil {
		return nil, errors.Join(fmt.Errorf("log exporter: %w", err), result.Shutdown(ctx))
	}

	result.logs = sdklog.NewLoggerProvider(sdklog.WithResource(res), sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)))
	otel.SetTracerProvider(result.traces)
	otel.SetMeterProvider(result.metrics)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	slog.SetDefault(slog.New(&LogHandler{Local: local, Remote: otelslog.NewHandler(Scope, otelslog.WithLoggerProvider(result.logs)), Level: level}))

	if err = registerRuntimeMetrics(result.metrics.Meter(Scope)); err != nil {
		return nil, errors.Join(err, result.Shutdown(ctx))
	}

	return result, nil
}

func (r *Runtime) Shutdown(ctx context.Context) error {
	if r.previous != nil {
		slog.SetDefault(r.previous)
	}

	var failures []error
	if r.traces != nil {
		failures = append(failures, r.traces.Shutdown(ctx))
	}

	if r.metrics != nil {
		failures = append(failures, r.metrics.Shutdown(ctx))
	}

	if r.logs != nil {
		failures = append(failures, r.logs.Shutdown(ctx))
	}

	return errors.Join(failures...)
}

func registerRuntimeMetrics(meter metric.Meter) error {
	goroutines, err := meter.Int64ObservableGauge("squadflow.runtime.goroutines")
	if err != nil {
		return erring.Wrap(err, "telemetry.runtime")
	}

	heap, err := meter.Float64ObservableGauge("squadflow.runtime.heap", metric.WithUnit("By"))
	if err != nil {
		return erring.Wrap(err, "telemetry.runtime")
	}

	_, err = meter.RegisterCallback(func(_ context.Context, observer metric.Observer) error {
		var stats runtime.MemStats

		runtime.ReadMemStats(&stats)
		observer.ObserveInt64(goroutines, int64(runtime.NumGoroutine()))
		observer.ObserveFloat64(heap, float64(stats.HeapAlloc))

		return nil
	}, goroutines, heap)

	return erring.Wrap(err, "telemetry.runtime")
}
