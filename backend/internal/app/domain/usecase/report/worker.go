package report

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

// Worker uses the durable job lease to fence duplicate deliveries and stale workers.
type Worker struct {
	Repository  JobRepository
	Reports     UseCaseInterface
	Queue       EventQueue
	Timeout     time.Duration
	Concurrency int
}

func (w *Worker) Process(ctx context.Context, event QueueEvent) error {
	if event.Type != EventRequested {
		return nil
	}

	token, err := identifier.NewUUID()
	if err != nil {
		return erring.Wrap(err, "report.worker")
	}

	timeout := w.Timeout
	if timeout <= 0 {
		timeout = 210 * time.Second
	}

	job, err := w.Repository.Claim(ctx, event.JobID, token, time.Now().Add(timeout+30*time.Second))
	if errors.Is(err, ErrLeaseLost) {
		return nil
	}

	if err != nil {
		return erring.Wrap(err, "report.worker")
	}

	workCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	workCtx = ctxkey.PutRequestID(workCtx, job.ID)
	workCtx = ctxkey.PutTransport(workCtx, "redis-stream")
	workCtx = ctxkey.PutUser(workCtx, entity.User{ID: job.UserID})

	var progressErr error

	result, buildErr := w.Reports.Build(workCtx, job.UserID, job.ID, job.Request, func(agent, message string) {
		if progressErr == nil {
			progressErr = w.Repository.Progress(workCtx, job, token, agent, message)
			if progressErr != nil {
				cancel()
			}
		}
	})

	if ctx.Err() != nil {
		return erring.Wrap(ctx.Err(), "report.worker.context")
	}

	if progressErr != nil {
		return erring.Wrap(progressErr, "report.worker.progress")
	}

	if buildErr == nil {
		return erring.Wrap(w.Repository.Complete(ctx, job, token, result), "report.worker.Complete")
	}

	slog.ErrorContext(workCtx, "asynchronous report attempt failed", "job_id", job.ID, "attempt", job.Attempts, "error", buildErr)

	message := "Não foi possível gerar o relatório. Se o problema persistir, contate o administrador."
	if errors.Is(buildErr, context.DeadlineExceeded) {
		message = "A geração excedeu o tempo disponível nesta tentativa."
	}

	delay := time.Duration(job.Attempts*job.Attempts) * 5 * time.Second

	return erring.Wrap(w.Repository.Fail(ctx, job, token, message, time.Now().Add(delay)), "report.worker.Fail")
}

func (w *Worker) Dispatch(ctx context.Context) error {
	token, err := identifier.NewUUID()
	if err != nil {
		return erring.Wrap(err, "report.worker")
	}

	events, err := w.Repository.ClaimEvents(ctx, token, time.Now().Add(15*time.Second))
	if err != nil {
		return erring.Wrap(err, "report.worker")
	}

	for _, event := range events {
		if err = w.Queue.Publish(ctx, event); err != nil {
			return erring.Wrap(err, "report.worker")
		}

		if err = w.Repository.Published(ctx, event.ID, token); err != nil {
			return erring.Wrap(err, "report.worker")
		}
	}

	return nil
}

// Run is independent of HTTP connections. Recovery also republishes work lost by the broker.
func (w *Worker) Run(ctx context.Context) {
	var running sync.WaitGroup

	count := w.Concurrency
	if count < 1 {
		count = 2
	}

	for range count {
		running.Add(1)

		go func() {
			defer running.Done()

			for ctx.Err() == nil {
				if err := w.Queue.Consume(ctx, w.Process); err != nil && ctx.Err() == nil {
					slog.ErrorContext(ctx, "report queue consumer retrying", "error", err)
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
				}
			}
		}()
	}

	running.Add(1)

	go func() {
		defer running.Done()

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			if err := w.Dispatch(ctx); err != nil && ctx.Err() == nil {
				slog.WarnContext(ctx, "report outbox pending", "error", err)
			}

			if err := w.Repository.Recover(ctx); err != nil && ctx.Err() == nil {
				slog.ErrorContext(ctx, "report recovery failed", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	running.Wait()
}
