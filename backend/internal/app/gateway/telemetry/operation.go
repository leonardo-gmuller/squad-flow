package telemetry

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	outcomeSuccess = "success"
	outcomeError   = "error"
)

type Operation struct {
	duration metric.Float64Histogram
	count    metric.Int64Counter
}

func NewOperation(name string) *Operation {
	meter := otel.Meter(Scope)
	duration, _ := meter.Float64Histogram("squadflow."+name+".duration", metric.WithUnit("s"), metric.WithExplicitBucketBoundaries(.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60, 120, 210))
	count, _ := meter.Int64Counter("squadflow." + name + ".operations")

	return &Operation{duration: duration, count: count}
}

func (o *Operation) Observe(ctx context.Context, start time.Time, outcome string, attrs ...attribute.KeyValue) {
	attrs = append(attrs, attribute.String("outcome", outcome))
	options := metric.WithAttributes(attrs...)
	o.duration.Record(ctx, time.Since(start).Seconds(), options)
	o.count.Add(ctx, 1, options)
}

// Failure records a bounded classification, never raw error messages or payloads.
func Failure(span trace.Span, err error) {
	if err == nil {
		return
	}

	kind := "operation_failed"
	if errors.Is(err, context.Canceled) {
		kind = "cancelled"
	}

	if errors.Is(err, context.DeadlineExceeded) {
		kind = "deadline_exceeded"
	}

	span.SetStatus(codes.Error, kind)
	span.RecordError(errors.New(kind))
}

func Inject(ctx context.Context) (string, string) {
	carrier := propagation.MapCarrier{}
	propagation.TraceContext{}.Inject(ctx, carrier)

	state := carrier.Get("tracestate")
	if len(state) > 512 {
		state = ""
	}

	return carrier.Get("traceparent"), state
}

func Extract(ctx context.Context, parent, state string) context.Context {
	if len(state) > 512 {
		state = ""
	}

	if len(parent) > 512 {
		parent = ""
	}

	return (propagation.TraceContext{}).Extract(ctx, propagation.MapCarrier{"traceparent": parent, "tracestate": state})
}
