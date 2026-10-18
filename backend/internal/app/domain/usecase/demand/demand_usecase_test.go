package demand

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

type transactionFixture struct {
	service     *UseCase
	demands     *MockRepository
	tests       *MockTestRepository
	transitions *MockTransitionRepository
	now         time.Time
}

func newTransactionFixture(t *testing.T) transactionFixture {
	t.Helper()
	ctrl := gomock.NewController(t)
	uow := NewMockUnitOfWork(ctrl)
	repos := NewMockRepositories(ctrl)
	fixture := transactionFixture{
		service: NewUseCase(uow, NewMockQueryRepository(ctrl)),
		demands: NewMockRepository(ctrl), tests: NewMockTestRepository(ctrl), transitions: NewMockTransitionRepository(ctrl),
		now: time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC),
	}
	fixture.service.now = func() time.Time { return fixture.now }
	// Accessors only expose the ports participating in this transaction.
	repos.EXPECT().Demands().Return(fixture.demands).AnyTimes()
	repos.EXPECT().Tests().Return(fixture.tests).AnyTimes()
	repos.EXPECT().Transitions().Return(fixture.transitions).AnyTimes()
	uow.EXPECT().Do(t.Context(), gomock.Any()).DoAndReturn(func(ctx context.Context, callback func(context.Context, Repositories) error) error {
		return callback(ctx, repos)
	})

	return fixture
}

func TestMoveEnforcesWorkflowAndCurrentQACycle(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name           string
		from, to       entity.Status
		passed, failed int
		expectedError  error
	}{
		{name: "cannot skip workflow", from: entity.Backlog, to: entity.Done, expectedError: erring.ErrInvalidTransition},
		{name: "completion requires tests", from: entity.InQA, to: entity.Done, expectedError: erring.ErrConflict},
		{name: "failure prevents completion", from: entity.InQA, to: entity.Done, passed: 1, failed: 1, expectedError: erring.ErrConflict},
		{name: "approved current cycle completes", from: entity.InQA, to: entity.Done, passed: 1},
		{name: "resubmission starts new cycle", from: entity.InProgress, to: entity.InQA},
		{name: "rework preserves counters", from: entity.ChangesRequested, to: entity.InProgress},
		{name: "rejection increments returns", from: entity.InQA, to: entity.ChangesRequested},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fixture := newTransactionFixture(t)
			item := entity.Demand{ID: "11111111-1111-4111-8111-111111111111", Status: test.from, QACycle: 2, ReturnCount: 1}
			fixture.demands.EXPECT().GetForUpdate(t.Context(), item.ID).Return(item, nil)

			if test.from == entity.InQA && test.to == entity.Done {
				fixture.tests.EXPECT().CurrentStats(t.Context(), item.ID, 2).Return(test.passed, test.failed, nil)
			}

			if test.expectedError == nil {
				expected := item
				expected.Status = test.to
				expected.UpdatedAt = fixture.now

				if test.to == entity.InQA {
					expected.QACycle = 3
				}

				if test.to == entity.ChangesRequested {
					expected.ReturnCount = 2
				}

				if test.to == entity.Done {
					expected.CompletedAt = &fixture.now
				}

				gomock.InOrder(
					fixture.demands.EXPECT().Update(t.Context(), expected).Return(nil),
					fixture.transitions.EXPECT().Insert(t.Context(), gomock.Cond(func(event entity.Transition) bool {
						return event.DemandID == item.ID && event.From == test.from && event.To == test.to && event.Actor == "Developer" && event.Reason == "Motivo" && event.CreatedAt.Equal(fixture.now)
					})).Return(nil),
				)
			}

			result, err := fixture.service.Move(t.Context(), item.ID, MoveInput{Status: test.to, Actor: " Developer ", Reason: " Motivo "})
			if !errors.Is(err, test.expectedError) {
				t.Fatalf("expected %v, received %v", test.expectedError, err)
			}

			if err == nil && result.Status != test.to {
				t.Fatalf("incorrect status: %+v", result)
			}
		})
	}
}

func TestMovePropagatesStorageFailures(t *testing.T) {
	t.Parallel()

	for _, stage := range []string{"lock", "stats", "update", "transition"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			fixture := newTransactionFixture(t)
			item := entity.Demand{ID: "11111111-1111-4111-8111-111111111111", Status: entity.InQA, QACycle: 2}
			storageError := errors.New("storage unavailable")

			var lockError, statsError, updateError, transitionError error

			switch stage {
			case "lock":
				lockError = storageError
			case "stats":
				statsError = storageError
			case "update":
				updateError = storageError
			case "transition":
				transitionError = storageError
			}

			fixture.demands.EXPECT().GetForUpdate(t.Context(), item.ID).Return(item, lockError)

			if lockError == nil {
				fixture.tests.EXPECT().CurrentStats(t.Context(), item.ID, 2).Return(1, 0, statsError)
			}

			if lockError == nil && statsError == nil {
				fixture.demands.EXPECT().Update(t.Context(), gomock.Any()).Return(updateError)
			}

			if lockError == nil && statsError == nil && updateError == nil {
				fixture.transitions.EXPECT().Insert(t.Context(), gomock.Any()).Return(transitionError)
			}

			result, err := fixture.service.Move(t.Context(), item.ID, MoveInput{Status: entity.Done})
			if !errors.Is(err, storageError) || result.ID != "" {
				t.Fatalf("failed operation returned a changed demand: %+v error=%v", result, err)
			}
		})
	}
}

func TestAddTestUsesCurrentCycleAndReturnsFailedDemandToDevelopment(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, result  string
		status        entity.Status
		expectedError error
	}{
		{name: "outside QA", result: "passed", status: entity.InProgress, expectedError: erring.ErrConflict},
		{name: "passed", result: "passed", status: entity.InQA},
		{name: "failed", result: "failed", status: entity.InQA},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fixture := newTransactionFixture(t)
			item := entity.Demand{ID: "11111111-1111-4111-8111-111111111111", Status: test.status, QACycle: 2, ReturnCount: 1}
			fixture.demands.EXPECT().GetForUpdate(t.Context(), item.ID).Return(item, nil)

			if test.expectedError == nil {
				insert := fixture.tests.EXPECT().Insert(t.Context(), gomock.Cond(func(record entity.TestCase) bool {
					return record.DemandID == item.ID && record.Cycle == 2 && record.Result == test.result && record.Author == "QA"
				})).Return(nil)

				if test.result == "failed" {
					expected := item
					expected.Status = entity.ChangesRequested
					expected.ReturnCount = 2
					expected.UpdatedAt = fixture.now
					update := fixture.demands.EXPECT().Update(t.Context(), expected).Return(nil).After(insert)
					fixture.transitions.EXPECT().Insert(t.Context(), gomock.Cond(func(event entity.Transition) bool {
						return event.DemandID == item.ID && event.From == entity.InQA && event.To == entity.ChangesRequested && event.Actor == "QA"
					})).Return(nil).After(update)
				}
			}

			record, err := fixture.service.AddTest(t.Context(), item.ID, AddTestInput{Title: "Validação do XML", Result: test.result, Actual: "Campo obrigatório ausente", Author: " QA "})
			if !errors.Is(err, test.expectedError) {
				t.Fatalf("expected %v, received %v", test.expectedError, err)
			}

			if err == nil && record.Cycle != 2 {
				t.Fatalf("incorrect QA cycle: %d", record.Cycle)
			}
		})
	}
}
