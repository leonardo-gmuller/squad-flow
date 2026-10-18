package demand

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

func TestMalformedResourceUUIDIsRejectedBeforePersistence(t *testing.T) {
	t.Parallel()

	for _, resourceID := range []string{"123", "invalid", "11111111111141118111111111111111"} {
		t.Run(resourceID, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			service := NewUseCase(NewMockUnitOfWork(ctrl), NewMockQueryRepository(ctrl))
			_, detailErr := service.Detail(t.Context(), resourceID)
			_, editErr := service.Edit(t.Context(), resourceID, EditInput{})
			_, moveErr := service.Move(t.Context(), resourceID, MoveInput{Status: entity.InProgress})
			_, testErr := service.AddTest(t.Context(), resourceID, AddTestInput{Title: "Valid test", Result: "passed"})
			_, createErr := service.Create(t.Context(), CreateInput{Title: "Valid demand title", Kind: entity.Improvement, Priority: "medium", PackageID: resourceID})

			for _, err := range []error{detailErr, editErr, moveErr, testErr, createErr} {
				if !errors.Is(err, erring.ErrInvalidInput) {
					t.Fatalf("malformed UUID reached persistence: %v", err)
				}
			}
		})
	}
}
