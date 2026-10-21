package telemetry

import (
	"context"
	"errors"
	"log/slog"

	"go.opentelemetry.io/otel/trace"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
)

// LogHandler keeps JSON stdout and OTLP logs correlated using the current context.
type LogHandler struct {
	Local  slog.Handler
	Remote slog.Handler
	Level  slog.Level
	bound  map[string]bool
}

func (h *LogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.Level && (h.Local.Enabled(ctx, level) || (h.Remote != nil && h.Remote.Enabled(ctx, level)))
}

func (h *LogHandler) Handle(ctx context.Context, record slog.Record) error {
	present := make(map[string]bool, len(h.bound)+record.NumAttrs())
	for key := range h.bound {
		present[key] = true
	}

	record.Attrs(func(attr slog.Attr) bool {
		present[attr.Key] = true
		return true
	})

	span := trace.SpanContextFromContext(ctx)
	if span.IsValid() {
		record.AddAttrs(slog.String("trace_id", span.TraceID().String()), slog.String("span_id", span.SpanID().String()))
	}

	if id, ok := ctxkey.GetRequestID(ctx); ok && !present["request_id"] {
		record.AddAttrs(slog.String("request_id", id))
	}

	if agent, ok := ctxkey.GetAgent(ctx); ok && !present["agent"] {
		record.AddAttrs(slog.String("agent", agent))
	}

	var localError, remoteError error
	if h.Local.Enabled(ctx, record.Level) {
		localError = h.Local.Handle(ctx, record.Clone())
	}

	if h.Remote != nil && h.Remote.Enabled(ctx, record.Level) {
		remoteError = h.Remote.Handle(ctx, record.Clone())
	}

	return errors.Join(localError, remoteError)
}

func (h *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	result := *h
	result.bound = make(map[string]bool, len(h.bound)+len(attrs))

	for key := range h.bound {
		result.bound[key] = true
	}

	for _, attr := range attrs {
		result.bound[attr.Key] = true
	}

	result.Local = h.Local.WithAttrs(attrs)

	if h.Remote != nil {
		result.Remote = h.Remote.WithAttrs(attrs)
	}

	return &result
}

func (h *LogHandler) WithGroup(name string) slog.Handler {
	result := *h
	result.Local = h.Local.WithGroup(name)

	if h.Remote != nil {
		result.Remote = h.Remote.WithGroup(name)
	}

	return &result
}
