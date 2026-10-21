//go:build integration

package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/redis/go-redis/v9"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	redisqueue "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/queue/redis"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/telemetry"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/testinfra"
)

//nolint:paralleltest // Sets a process-wide tracer provider; other integration tests remain suspended.
func TestObservabilityTraceSurvivesOutboxRedisAndWorker(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()), sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(provider)

	defer func() {
		_ = provider.Shutdown(t.Context())

		otel.SetTracerProvider(tracenoop.NewTracerProvider())
	}()

	_, service, _, user := reportJobFixture(t)
	url := testinfra.Redis(t)

	options, err := redis.ParseURL(url)
	if err != nil {
		t.Fatal(err)
	}

	raw := redis.NewClient(options)

	t.Cleanup(func() { _ = raw.Close() })

	queue, err := redisqueue.New(url, "observability-reports", "observability-workers", time.Second)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = queue.Close() })

	instrumented := telemetry.Events(queue)
	ctx, cancelRequest := context.WithCancel(t.Context())
	ctx, requestSpan := otel.Tracer(telemetry.Scope).Start(ctx, "submit-report")
	requestTrace := requestSpan.SpanContext().TraceID()

	job, err := service.Submit(ctx, user, report.Request{Prompt: "Analise o fluxo de entregas"}, "")
	if err != nil {
		t.Fatal(err)
	}

	requestSpan.End()
	cancelRequest()

	if job.TraceParent == "" {
		t.Fatal("initial trace context was not persisted")
	}

	encoded, err := json.Marshal(job)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(encoded), "trace_parent") || strings.Contains(string(encoded), job.TraceParent) {
		t.Fatal("internal trace carrier leaked in the job API representation")
	}

	token, err := identifier.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	events, err := service.Repository.ClaimEvents(t.Context(), token, time.Now().Add(time.Minute))
	if err != nil || len(events) != 1 {
		t.Fatalf("outbox claim: events=%v error=%v", events, err)
	}

	if events[0].TraceParent != job.TraceParent {
		t.Fatal("trace context was not restored from the outbox")
	}

	if err = instrumented.Publish(t.Context(), events[0]); err != nil {
		t.Fatal(err)
	}

	if err = service.Repository.Published(t.Context(), events[0].ID, token); err != nil {
		t.Fatal(err)
	}

	worker := report.Worker{Repository: telemetry.Jobs(service.Repository), Reports: telemetry.Reports(service.Reports), Timeout: 5 * time.Second}

	workerCtx, cancelWorker := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancelWorker()

	results := make(chan error, 1)
	processed := make(chan struct{})

	go func() {
		results <- instrumented.Consume(workerCtx, func(ctx context.Context, event report.QueueEvent) error {
			if trace.SpanContextFromContext(ctx).TraceID() != requestTrace {
				return errors.New("consumer lost request trace")
			}
			if processError := worker.Process(ctx, event); processError != nil {
				return erring.Wrap(processError, "observability.test")
			}
			close(processed)

			return nil
		})
	}()
	select {
	case <-processed:
	case err = <-results:
		t.Fatalf("consumer stopped: %v", err)
	case <-workerCtx.Done():
		t.Fatal("worker did not process event")
	}
	waitForObservedAck(workerCtx, t, raw, "observability-reports")
	cancelWorker()
	select {
	case <-results:
	case <-time.After(3 * time.Second):
		t.Fatal("consumer did not stop")
	}

	saved, err := service.Get(t.Context(), user.ID, job.ID)
	if err != nil || saved.Status != report.JobCompleted {
		t.Fatalf("job not completed: %+v %v", saved, err)
	}

	assertObservedTrace(t, recorder, requestTrace)
}

func waitForObservedAck(ctx context.Context, t *testing.T, raw *redis.Client, stream string) {
	t.Helper()

	// Wait for ACK/delete before cancellation, so the broker path is verified too.
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	for {
		length, lengthError := raw.XLen(ctx, stream).Result()
		if lengthError != nil {
			t.Fatal(lengthError)
		}

		if length == 0 {
			break
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			t.Fatal("event was not acknowledged")
		}
	}
}

func assertObservedTrace(t *testing.T, recorder *tracetest.SpanRecorder, requestTrace trace.TraceID) {
	t.Helper()

	spans := map[string]sdktrace.ReadOnlySpan{}

	for _, span := range recorder.Ended() {
		if span.SpanContext().TraceID() == requestTrace {
			spans[span.Name()] = span
		}
	}

	for _, name := range []string{"submit-report", "report.requested publish", "report.requested process", "report.build"} {
		if spans[name] == nil {
			t.Errorf("missing correlated span: %s", name)
		}
	}

	if spans["report.requested process"] != nil && spans["report.requested publish"] != nil && spans["report.requested process"].Parent().SpanID() != spans["report.requested publish"].SpanContext().SpanID() {
		t.Fatal("Redis consumer lost producer parent")
	}
}
