package report

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
)

func TestReportScopeValidationChecksOwnershipAndCompletedDemand(t *testing.T) {
	t.Parallel()

	ownerID := "11111111-1111-4111-8111-111111111111"
	demandID := "22222222-2222-4222-8222-222222222222"
	previousID := "33333333-3333-4333-8333-333333333333"

	for _, test := range []struct {
		name          string
		previousError error
		status        entity.Status
		expectedError error
	}{
		{name: "previous report belongs to another user", previousError: erring.ErrNotFound, expectedError: erring.ErrNotFound},
		{name: "demand is still in QA", status: entity.InQA, expectedError: erring.ErrConflict},
		{name: "completed demand and owned report", status: entity.Done},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			repository := NewMockRepository(ctrl)
			data := NewMockData(ctrl)
			previous := repository.EXPECT().Get(t.Context(), ownerID, previousID).Return(Report{ID: previousID}, test.previousError)

			if test.previousError == nil {
				data.EXPECT().Detail(t.Context(), demandID).Return(demand.Detail{Demand: entity.Demand{ID: demandID, Status: test.status}}, nil).After(previous)
			}

			service := &Service{Repository: repository, Data: data, Generator: NewMockGenerator(ctrl)}

			err := service.Validate(t.Context(), ownerID, Request{Prompt: "Analise as entregas", DemandID: demandID, PreviousReportID: previousID})
			if !errors.Is(err, test.expectedError) {
				t.Fatalf("expected %v, received %v", test.expectedError, err)
			}
		})
	}
}

func TestReportQueriesRejectMalformedUUIDsBeforePersistence(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	service := &Service{Repository: NewMockRepository(ctrl), Data: NewMockData(ctrl)}
	_, getError := service.Get(t.Context(), "owner", "invalid")
	_, listError := service.List(t.Context(), "owner", "invalid")
	validateError := service.Validate(t.Context(), "owner", Request{Prompt: "Analise as entregas", DemandID: "invalid"})

	for _, err := range []error{getError, listError, validateError} {
		if !errors.Is(err, erring.ErrInvalidInput) {
			t.Fatalf("malformed UUID reached persistence: %v", err)
		}
	}
}
