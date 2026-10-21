package telemetry

import (
	"strings"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestTracePropagationBoundsTracestateAndPreservesTrace(t *testing.T) {
	t.Parallel()

	parent := "00-11111111111111111111111111111111-2222222222222222-01"
	state := "a=" + strings.Repeat("x", 250) + ",b=" + strings.Repeat("y", 250) + ",c=value"
	ctx := Extract(t.Context(), parent, state)
	encoded, encodedState := Inject(ctx)

	if encoded != parent || encodedState != "" || !trace.SpanContextFromContext(ctx).IsValid() {
		t.Fatal("trace context lost or oversized metadata was propagated")
	}

	invalid := Extract(t.Context(), "invalid", "")
	if trace.SpanContextFromContext(invalid).IsValid() {
		t.Fatal("malformed traceparent accepted")
	}
}

func TestPostgresSpanNamesDoNotContainSQLOrParameters(t *testing.T) {
	t.Parallel()

	for _, test := range []struct{ sql, name string }{
		{"-- name: GetUser :one\nSELECT password_hash FROM users WHERE email=$1", "GetUser"},
		{"SELECT 'private-value'", "statement"},
		{"-- name: bad'name :one", "statement"},
		{"BEGIN", "begin"},
	} {
		if name := queryName(test.sql); name != test.name {
			t.Fatalf("unsafe query name: %q", name)
		}
	}
}
