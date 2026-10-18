package demand

import (
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
)

func TestEditDoesNotResetQAHistory(t *testing.T) {
	t.Parallel()
	fixture := newTransactionFixture(t)
	original := entity.Demand{ID: "11111111-1111-4111-8111-111111111111", Title: "Descrição original", Kind: entity.Improvement, Priority: "medium", Status: entity.InProgress, QACycle: 2, ReturnCount: 1, CreatedAt: fixture.now.AddDate(0, 0, -1)}
	expected := original
	expected.Title = "Nova descrição do item"
	expected.Priority = "high"
	expected.AssigneeID = "55555555-5555-4555-8555-555555555501"
	expected.ReviewerID = "55555555-5555-4555-8555-555555555502"
	expected.AcceptanceCriteria = "XML validado"
	expected.UpdatedAt = fixture.now
	expected.EventCodes = []string{}
	// No transition or test write is allowed while editing details.
	gomock.InOrder(
		fixture.demands.EXPECT().GetForUpdate(t.Context(), original.ID).Return(original, nil),
		fixture.demands.EXPECT().UpdateDetails(t.Context(), expected).Return(nil),
		fixture.demands.EXPECT().GetForUpdate(t.Context(), original.ID).Return(expected, nil),
	)

	updated, err := fixture.service.Edit(t.Context(), original.ID, EditInput{Title: " Nova descrição do item ", Priority: "high", AssigneeID: expected.AssigneeID, ReviewerID: expected.ReviewerID, AcceptanceCriteria: "XML validado"})
	if err != nil {
		t.Fatal(err)
	}

	if updated.Status != original.Status || updated.QACycle != 2 || updated.ReturnCount != 1 || !updated.CreatedAt.Equal(original.CreatedAt) {
		t.Fatalf("editing reset workflow state: %+v", updated)
	}
}
