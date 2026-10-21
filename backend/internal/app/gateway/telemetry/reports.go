package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
)

type ReportService struct {
	report.UseCaseInterface
	operation *Operation
}

func Reports(service report.UseCaseInterface) report.UseCaseInterface {
	return &ReportService{UseCaseInterface: service, operation: NewOperation("report.build")}
}

func (s *ReportService) Build(ctx context.Context, userID, reportID string, input report.Request, progress report.Progress) (report.Report, error) {
	ctx, span := otel.Tracer(Scope).Start(ctx, "report.build", trace.WithAttributes(attribute.String("report.job.id", reportID)))
	defer span.End()

	start := time.Now()
	result, err := s.UseCaseInterface.Build(ctx, userID, reportID, input, func(agent, message string) {
		// Messages and generated content are intentionally omitted from telemetry.
		span.AddEvent("report.progress", trace.WithAttributes(attribute.String("report.stage", stageName(agent))))

		if progress != nil {
			progress(agent, message)
		}
	})
	outcome := outcomeSuccess

	if err != nil {
		outcome = outcomeError

		Failure(span, err)
	}

	s.operation.Observe(ctx, start, outcome)

	return result, erring.Wrap(err, "telemetry.report")
}

func stageName(agent string) string {
	switch agent {
	case "coleta", "análise", "revisão", "correção":
		return agent
	default:
		return "other"
	}
}

type Queue struct {
	report.EventQueue
	operation *Operation
}

func Events(queue report.EventQueue) report.EventQueue {
	return &Queue{EventQueue: queue, operation: NewOperation("queue")}
}

func (q *Queue) Publish(ctx context.Context, event report.QueueEvent) error {
	ctx = Extract(ctx, event.TraceParent, event.TraceState)

	ctx, span := otel.Tracer(Scope).Start(ctx, "report.requested publish", trace.WithSpanKind(trace.SpanKindProducer), trace.WithAttributes(attribute.String("messaging.system", "redis"), attribute.String("messaging.operation.name", "publish"), attribute.String("report.job.id", event.JobID)))
	defer span.End()

	event.TraceParent, event.TraceState = Inject(ctx)
	start := time.Now()
	err := q.EventQueue.Publish(ctx, event)
	outcome := outcomeSuccess

	if err != nil {
		outcome = outcomeError

		Failure(span, err)
	}

	q.operation.Observe(ctx, start, outcome, attribute.String("operation", "publish"))

	return erring.Wrap(err, "telemetry.operation")
}

func (q *Queue) Consume(ctx context.Context, handle func(context.Context, report.QueueEvent) error) error {
	consumeErr := q.EventQueue.Consume(ctx, func(ctx context.Context, event report.QueueEvent) error {
		ctx = Extract(ctx, event.TraceParent, event.TraceState)

		ctx, span := otel.Tracer(Scope).Start(ctx, "report.requested process", trace.WithSpanKind(trace.SpanKindConsumer), trace.WithAttributes(attribute.String("messaging.system", "redis"), attribute.String("messaging.operation.name", "process"), attribute.String("report.job.id", event.JobID)))
		defer span.End()

		start := time.Now()
		err := handle(ctx, event)
		outcome := outcomeSuccess

		if err != nil {
			outcome = outcomeError

			Failure(span, err)
		}

		q.operation.Observe(ctx, start, outcome, attribute.String("operation", "process"))

		return erring.Wrap(err, "telemetry.operation")
	})

	return erring.Wrap(consumeErr, "telemetry.consume")
}

// JobRepository observes actual committed outcomes, including scheduled retries.
type JobRepository struct{ report.JobRepository }

func Jobs(repository report.JobRepository) report.JobRepository {
	return &JobRepository{JobRepository: repository}
}

func (r *JobRepository) Complete(ctx context.Context, job report.Job, token string, result report.Report) error {
	err := r.JobRepository.Complete(ctx, job, token, result)
	if err == nil {
		recordJob(ctx, "completed")
	}

	return erring.Wrap(err, "telemetry.operation")
}

func (r *JobRepository) Fail(ctx context.Context, job report.Job, token, message string, next time.Time) error {
	err := r.JobRepository.Fail(ctx, job, token, message, next)
	if err == nil {
		outcome := "retry"
		if job.Attempts >= job.MaxAttempts {
			outcome = "failed"
		}

		recordJob(ctx, outcome)
	}

	return erring.Wrap(err, "telemetry.operation")
}

func recordJob(ctx context.Context, outcome string) {
	counter, _ := otel.Meter(Scope).Int64Counter("squadflow.report.jobs")
	counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", outcome)))
}
