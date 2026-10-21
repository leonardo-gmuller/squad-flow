package telemetry

import (
	"context"
	"strings"
	"time"
	"unicode"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

type (
	QueryTracer   struct{ operation *Operation }
	queryStartKey struct{}
)

func NewQueryTracer() *QueryTracer { return &QueryTracer{operation: NewOperation("db.query")} }

//nolint:spancheck // pgx ends the query span in TraceQueryEnd using the returned context.
func (t *QueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	name := queryName(data.SQL)
	ctx, _ = otel.Tracer(Scope).Start(ctx, "postgres "+name, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attribute.String("db.system.name", "postgresql"), attribute.String("db.operation.name", name)))

	return context.WithValue(ctx, queryStartKey{}, time.Now())
}

func (t *QueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)
	defer span.End()
	Failure(span, data.Err)

	start, ok := ctx.Value(queryStartKey{}).(time.Time)
	if ok {
		outcome := outcomeSuccess
		if data.Err != nil {
			outcome = outcomeError
		}

		t.operation.Observe(ctx, start, outcome)
	}
}

// Only SQLC operation names are exported. SQL text, arguments and DSNs stay private.
func queryName(sql string) string {
	first, _, _ := strings.Cut(sql, "\n")
	if strings.HasPrefix(first, "-- name: ") {
		fields := strings.Fields(first)
		if len(fields) >= 3 && len(fields[2]) <= 80 && strings.IndexFunc(fields[2], func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' }) == -1 {
			return fields[2]
		}
	}

	switch strings.ToUpper(strings.TrimSpace(first)) {
	case "BEGIN", "COMMIT", "ROLLBACK":
		return strings.ToLower(strings.TrimSpace(first))
	default:
		return "statement"
	}
}

func RegisterPool(pool *pgxpool.Pool) (metric.Registration, error) {
	meter := otel.Meter(Scope)

	total, err := meter.Int64ObservableGauge("squadflow.db.pool.connections")
	if err != nil {
		return nil, erring.Wrap(err, "telemetry.instrument")
	}

	acquired, err := meter.Int64ObservableGauge("squadflow.db.pool.acquired")
	if err != nil {
		return nil, erring.Wrap(err, "telemetry.instrument")
	}

	idle, err := meter.Int64ObservableGauge("squadflow.db.pool.idle")
	if err != nil {
		return nil, erring.Wrap(err, "telemetry.instrument")
	}

	registration, err := meter.RegisterCallback(func(_ context.Context, observer metric.Observer) error {
		stats := pool.Stat()
		observer.ObserveInt64(total, int64(stats.TotalConns()))
		observer.ObserveInt64(acquired, int64(stats.AcquiredConns()))
		observer.ObserveInt64(idle, int64(stats.IdleConns()))

		return nil
	}, total, acquired, idle)

	return registration, erring.Wrap(err, "telemetry.pool")
}
