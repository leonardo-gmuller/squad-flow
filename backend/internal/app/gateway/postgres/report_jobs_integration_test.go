//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres"
	authrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/auth"
	reportrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/report"
	reportjobrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/reportjob"
	redisqueue "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/queue/redis"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/reporthtml"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/testinfra"
)

type asyncGenerator struct {
	calls atomic.Int32
	fail  bool
}

func (g *asyncGenerator) Generate(_ context.Context, _ string, _ report.Toolset, progress report.Progress) (report.Content, error) {
	g.calls.Add(1)

	if g.fail {
		return report.Content{}, errors.New("simulated provider failure with private details")
	}

	progress("análise", "Conferindo indicadores")

	return report.Content{Title: "Análise assíncrona", Summary: "Documento pronto", HTML: `<html><head><style>h1{color:purple}</style></head><body><h1>Entregas</h1><p>Indicadores consultados.</p></body></html>`}, nil
}

func reportJobFixture(t *testing.T) (*postgres.Client, *report.JobService, *asyncGenerator, entity.User) {
	t.Helper()

	databaseURL := testinfra.Postgres(t)

	root, err := postgres.New(t.Context(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(root.Close)

	schemaID, _ := identifier.NewUUID()

	schema := "async_" + strings.ReplaceAll(schemaID, "-", "")
	if _, err = root.Pool.Exec(t.Context(), "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if _, cleanupErr := root.Pool.Exec(context.WithoutCancel(t.Context()), "DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE"); cleanupErr != nil {
			t.Error(cleanupErr)
		}
	})

	connectionURL, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	params := connectionURL.Query()
	params.Set("search_path", schema)
	connectionURL.RawQuery = params.Encode()

	db, err := postgres.New(t.Context(), connectionURL.String())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(db.Close)

	if err = db.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}

	id, err := identifier.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	user := entity.User{ID: id, Name: "Async report test", Email: id + "@example.test", PasswordHash: "unused-test-hash", CreatedAt: time.Now()}
	if err = authrepo.NewAuthRepository(db.Pool).CreateUser(t.Context(), user); err != nil {
		t.Fatal(err)
	}

	generator := &asyncGenerator{}
	service := &report.Service{Data: emptyReportData{}, Repository: reportrepo.NewReportRepository(db.Pool), Generator: generator, Renderer: reporthtml.Renderer{}}

	return db, &report.JobService{Repository: reportjobrepo.NewReportJobRepository(db.Pool), Reports: service}, generator, user
}

type emptyReportData struct{}

func (emptyReportData) List(context.Context, string, string, string) ([]entity.Demand, error) {
	return []entity.Demand{}, nil
}

func (emptyReportData) Detail(context.Context, string) (demand.Detail, error) {
	return demand.Detail{}, erring.ErrNotFound
}

func TestReportJobsPersistIdempotencyOwnershipAndDuplicateDelivery(t *testing.T) {
	t.Parallel()
	_, service, generator, user := reportJobFixture(t)
	key, _ := identifier.NewUUID()
	input := report.Request{Prompt: "Analise as entregas do workspace"}

	job, err := service.Submit(t.Context(), user, input, key)
	if err != nil {
		t.Fatal(err)
	}

	if generator.calls.Load() != 0 || job.Status != report.JobQueued {
		t.Fatal("submission executed the LLM")
	}

	replay, err := service.Submit(t.Context(), user, input, key)
	if err != nil || replay.ID != job.ID {
		t.Fatalf("idempotency lost: %+v %v", replay, err)
	}

	if _, err = service.Submit(t.Context(), user, report.Request{Prompt: "Outro pedido"}, key); !errors.Is(err, erring.ErrConflict) {
		t.Fatalf("idempotency key accepted different request: %v", err)
	}

	if _, err = service.Submit(t.Context(), user, input, ""); !errors.Is(err, erring.ErrConflict) {
		t.Fatalf("concurrent active job accepted: %v", err)
	}

	stranger, _ := identifier.NewUUID()
	if _, err = service.Get(t.Context(), stranger, job.ID); !errors.Is(err, erring.ErrNotFound) {
		t.Fatalf("foreign job exposed: %v", err)
	}

	events, err := service.Repository.History(t.Context(), user.ID, job.ID)
	if err != nil || len(events) != 1 {
		t.Fatalf("duplicate submission added events: %+v %v", events, err)
	}

	worker := report.Worker{Repository: service.Repository, Reports: service.Reports, Timeout: time.Second * 5}
	event := report.QueueEvent{ID: events[0].ID, JobID: job.ID, Type: report.EventRequested}

	for range 2 {
		if err = worker.Process(t.Context(), event); err != nil {
			t.Fatal(err)
		}
	}

	finished, err := service.Get(t.Context(), user.ID, job.ID)
	if err != nil {
		t.Fatal(err)
	}

	if finished.Status != report.JobCompleted || finished.Attempts != 1 || finished.ReportID != job.ID || generator.calls.Load() != 1 {
		t.Fatalf("duplicate event repeated generation: %+v", finished)
	}

	saved, err := service.Reports.Get(t.Context(), user.ID, job.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(saved.Content.HTML, "Content-Security-Policy") || len(saved.SourceData) == 0 {
		t.Fatal("HTML or evidence snapshot missing")
	}

	history, err := service.Reports.List(t.Context(), user.ID, "")
	if err != nil {
		t.Fatal(err)
	}

	if len(history) != 1 || history[0].Content.HTML != "" {
		t.Fatal("history list downloaded HTML documents eagerly")
	}
}

func TestReportJobsRecoverLeasesRetryAndFenceStaleWorkers(t *testing.T) {
	t.Parallel()
	db, service, generator, user := reportJobFixture(t)
	generator.fail = true

	job, err := service.Submit(t.Context(), user, report.Request{Prompt: "Analise os indicadores"}, "")
	if err != nil {
		t.Fatal(err)
	}

	worker := report.Worker{Repository: service.Repository, Reports: service.Reports, Timeout: 5 * time.Second}
	if err = worker.Process(t.Context(), report.QueueEvent{JobID: job.ID, Type: report.EventRequested}); err != nil {
		t.Fatal(err)
	}

	failed, err := service.Get(t.Context(), user.ID, job.ID)
	if err != nil {
		t.Fatal(err)
	}

	if failed.Status != report.JobQueued || failed.Attempts != 1 || strings.Contains(failed.ErrorMessage, "private") {
		t.Fatalf("retry not persisted safely: %+v", failed)
	}

	if _, err = db.Pool.Exec(t.Context(), `UPDATE report_jobs SET available_at=now()-interval '1 second' WHERE id=$1`, job.ID); err != nil {
		t.Fatal(err)
	}

	staleToken, _ := identifier.NewUUID()

	claimed, err := service.Repository.Claim(t.Context(), job.ID, staleToken, time.Now().Add(-time.Second))
	if err != nil {
		t.Fatal(err)
	}

	if err = service.Repository.Recover(t.Context()); err != nil {
		t.Fatal(err)
	}

	if err = service.Repository.Progress(t.Context(), claimed, staleToken, "análise", "stale"); !errors.Is(err, report.ErrLeaseLost) {
		t.Fatalf("stale worker updated progress: %v", err)
	}

	if err = service.Repository.Complete(t.Context(), claimed, staleToken, report.Report{ID: job.ID}); !errors.Is(err, report.ErrLeaseLost) {
		t.Fatalf("stale worker published report: %v", err)
	}

	if err = worker.Process(t.Context(), report.QueueEvent{JobID: job.ID, Type: report.EventRequested}); err != nil {
		t.Fatal(err)
	}

	terminal, err := service.Get(t.Context(), user.ID, job.ID)
	if err != nil {
		t.Fatal(err)
	}

	if terminal.Status != report.JobFailed || terminal.Attempts != 3 {
		t.Fatalf("retry limit ignored: %+v", terminal)
	}

	retried, err := service.Retry(t.Context(), user, job.ID, "")
	if err != nil {
		t.Fatal(err)
	}

	if retried.ID == job.ID || retried.Attempts != 0 || retried.Status != report.JobQueued {
		t.Fatal("manual retry destroyed the original history")
	}
	// A queued job whose Redis message disappeared gets a fresh outbox event.
	if _, err = db.Pool.Exec(t.Context(), `UPDATE report_jobs SET updated_at=now()-interval '61 seconds' WHERE id=$1`, retried.ID); err != nil {
		t.Fatal(err)
	}

	if err = service.Repository.Recover(t.Context()); err != nil {
		t.Fatal(err)
	}

	recovered, err := service.Get(t.Context(), user.ID, retried.ID)
	if err != nil || len(recovered.Events) != 2 {
		t.Fatalf("lost broker message was not recovered: %+v %v", recovered, err)
	}
}

func TestReportWorkerConsumesRedisAfterSubmissionContextEnds(t *testing.T) {
	t.Parallel()

	redisURL := testinfra.Redis(t)

	_, service, _, user := reportJobFixture(t)
	requestCtx, cancelRequest := context.WithCancel(t.Context())
	job, err := service.Submit(requestCtx, user, report.Request{Prompt: "Prepare uma apresentação"}, "")

	cancelRequest()

	if err != nil {
		t.Fatal(err)
	}

	stream := "squadflow:test:reports:" + job.ID

	options, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatal(err)
	}

	cleanupRedis := redis.NewClient(options)

	t.Cleanup(func() {
		_ = cleanupRedis.Del(context.WithoutCancel(t.Context()), stream).Err()
		_ = cleanupRedis.Close()
	})

	queue, err := redisqueue.New(redisURL, stream, "test-workers", time.Second)
	if err != nil {
		t.Fatal(err)
	}

	defer queue.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	worker := report.Worker{Repository: service.Repository, Reports: service.Reports, Queue: queue, Timeout: 5 * time.Second, Concurrency: 2}

	go func() { defer close(done); worker.Run(ctx) }()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		state, getErr := service.Get(ctx, user.ID, job.ID)
		if getErr != nil {
			t.Fatal(getErr)
		}

		if state.Status == report.JobCompleted {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("worker failed to complete durable job")
		case <-ticker.C:
		}
	}

	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("worker did not stop")
	}
}

func TestReportJobClaimIsExclusive(t *testing.T) {
	t.Parallel()
	_, service, _, user := reportJobFixture(t)

	job, err := service.Submit(t.Context(), user, report.Request{Prompt: "Analise o desempenho"}, "")
	if err != nil {
		t.Fatal(err)
	}

	results := make(chan error, 2)

	for range 2 {
		go func() {
			token, _ := identifier.NewUUID()
			_, claimErr := service.Repository.Claim(t.Context(), job.ID, token, time.Now().Add(time.Minute))
			results <- claimErr
		}()
	}

	claims := 0

	for range 2 {
		claimErr := <-results
		if claimErr == nil {
			claims++
		} else if !errors.Is(claimErr, report.ErrLeaseLost) {
			t.Fatal(claimErr)
		}
	}

	if claims != 1 {
		t.Fatalf("job claimed %d times", claims)
	}

	state, err := service.Get(t.Context(), user.ID, job.ID)
	if err != nil || state.Attempts != 1 {
		t.Fatalf("duplicate attempt recorded: %+v %v", state, err)
	}
}

type outboxProbe struct {
	fail      bool
	published []report.QueueEvent
}

func (p *outboxProbe) Publish(_ context.Context, event report.QueueEvent) error {
	if p.fail {
		return errors.New("simulated broker outage")
	}

	p.published = append(p.published, event)

	return nil
}

func (*outboxProbe) Consume(context.Context, func(context.Context, report.QueueEvent) error) error {
	return nil
}

func TestReportOutboxSurvivesBrokerOutage(t *testing.T) {
	t.Parallel()
	db, service, _, user := reportJobFixture(t)

	job, err := service.Submit(t.Context(), user, report.Request{Prompt: "Analise a qualidade"}, "")
	if err != nil {
		t.Fatal(err)
	}

	queue := &outboxProbe{fail: true}

	worker := report.Worker{Repository: service.Repository, Reports: service.Reports, Queue: queue}
	if err = worker.Dispatch(t.Context()); err == nil {
		t.Fatal("simulated outage did not fail")
	}

	var published bool
	if err = db.Pool.QueryRow(t.Context(), `SELECT published_at IS NOT NULL FROM report_events WHERE job_id=$1 AND event_type='report.requested'`, job.ID).Scan(&published); err != nil {
		t.Fatal(err)
	}

	if published {
		t.Fatal("failed delivery was marked as published")
	}

	if _, err = db.Pool.Exec(t.Context(), `UPDATE report_events SET publish_lease_until=now()-interval '1 second' WHERE job_id=$1`, job.ID); err != nil {
		t.Fatal(err)
	}

	queue.fail = false

	if err = worker.Dispatch(t.Context()); err != nil {
		t.Fatal(err)
	}

	if len(queue.published) != 1 || queue.published[0].JobID != job.ID {
		t.Fatal("outbox event lost after broker recovery")
	}

	if err = worker.Dispatch(t.Context()); err != nil {
		t.Fatal(err)
	}

	if len(queue.published) != 1 {
		t.Fatal("already published event was dispatched again")
	}
}
