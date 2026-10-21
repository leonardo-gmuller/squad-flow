package report

import (
	"context"
	"errors"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

const (
	JobQueued      = "queued"
	JobProcessing  = "processing"
	JobCompleted   = "completed"
	JobFailed      = "failed"
	EventRequested = "report.requested"
)

var ErrLeaseLost = errors.New("report processing lease is no longer owned")

type Job struct {
	TraceParent     string     `json:"-"`
	TraceState      string     `json:"-"`
	ID              string     `json:"id"`
	UserID          string     `json:"-"`
	RequestKey      string     `json:"-"`
	Request         Request    `json:"request"`
	Status          string     `json:"status"`
	Attempts        int        `json:"attempts"`
	MaxAttempts     int        `json:"max_attempts"`
	AvailableAt     time.Time  `json:"available_at"`
	ProgressAgent   string     `json:"progress_agent"`
	ProgressMessage string     `json:"progress_message"`
	ErrorMessage    string     `json:"error_message"`
	ReportID        string     `json:"report_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	Events          []JobEvent `json:"events,omitempty"`
}
type JobEvent struct {
	ID        string    `json:"id"`
	JobID     string    `json:"job_id"`
	Type      string    `json:"type"`
	Agent     string    `json:"agent,omitempty"`
	Message   string    `json:"message,omitempty"`
	Attempt   int       `json:"attempt"`
	CreatedAt time.Time `json:"created_at"`
}
type QueueEvent struct {
	TraceParent string `json:"-"`
	TraceState  string `json:"-"`
	ID          string `json:"id"`
	JobID       string `json:"job_id"`
	Type        string `json:"type"`
}

type JobRequests interface {
	Enqueue(ctx context.Context, job Job) (Job, error)
	Get(ctx context.Context, userID, id string) (Job, error)
	List(ctx context.Context, userID string) ([]Job, error)
	History(ctx context.Context, userID, id string) ([]JobEvent, error)
}
type JobExecution interface {
	Claim(ctx context.Context, id, token string, until time.Time) (Job, error)
	Progress(ctx context.Context, job Job, token, agent, message string) error
	Complete(ctx context.Context, job Job, token string, result Report) error
	Fail(ctx context.Context, job Job, token, message string, next time.Time) error
}
type JobOutbox interface {
	ClaimEvents(ctx context.Context, token string, until time.Time) ([]QueueEvent, error)
	Published(ctx context.Context, id, token string) error
	Recover(ctx context.Context) error
}
type JobRepository interface {
	JobRequests
	JobExecution
	JobOutbox
}
type EventQueue interface {
	Publish(ctx context.Context, event QueueEvent) error
	Consume(ctx context.Context, handle func(ctx context.Context, event QueueEvent) error) error
}
type DocumentRenderer interface {
	Render(content Content) (string, error)
}

type JobService struct {
	Repository JobRepository
	Reports    UseCaseInterface
}

func (s *JobService) Submit(ctx context.Context, user entity.User, input Request, requestKey string) (Job, error) {
	if !identifier.IsUUID(user.ID) {
		return Job{}, erring.ErrForbidden
	}

	if err := s.Reports.Validate(ctx, user.ID, input); err != nil {
		return Job{}, erring.Wrap(err, "report.jobs")
	}

	if requestKey == "" {
		var err error

		requestKey, err = identifier.NewUUID()
		if err != nil {
			return Job{}, erring.Wrap(err, "report.jobs")
		}
	}

	if !identifier.IsUUID(requestKey) {
		return Job{}, erring.ErrInvalidInput.WithMessage("Idempotency-Key deve ser um UUID")
	}

	id, err := identifier.NewUUID()
	if err != nil {
		return Job{}, erring.Wrap(err, "report.jobs")
	}

	job, err := s.Repository.Enqueue(ctx, Job{ID: id, UserID: user.ID, RequestKey: requestKey, Request: input, MaxAttempts: 3})

	return job, erring.Wrap(err, "report.jobs.Enqueue")
}

func (s *JobService) Get(ctx context.Context, userID, id string) (Job, error) {
	if !identifier.IsUUID(id) {
		return Job{}, erring.ErrInvalidInput
	}

	job, err := s.Repository.Get(ctx, userID, id)
	if err != nil {
		return Job{}, erring.Wrap(err, "report.jobs")
	}

	job.Events, err = s.Repository.History(ctx, userID, id)

	return job, erring.Wrap(err, "report.jobs.Get")
}

func (s *JobService) Retry(ctx context.Context, user entity.User, id, key string) (Job, error) {
	job, err := s.Get(ctx, user.ID, id)
	if err != nil {
		return Job{}, erring.Wrap(err, "report.jobs")
	}

	if job.Status != JobFailed {
		return Job{}, erring.ErrConflict.WithMessage("somente solicitações com falha podem ser reenviadas")
	}

	return s.Submit(ctx, user, job.Request, key)
}

func (s *JobService) List(ctx context.Context, userID string) ([]Job, error) {
	jobs, err := s.Repository.List(ctx, userID)

	return jobs, erring.Wrap(err, "report.jobs.List")
}
