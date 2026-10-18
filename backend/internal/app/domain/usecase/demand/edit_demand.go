package demand

import (
	"context"
	"strings"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

// Edits do not change the workflow state, QA history or metrics counters.
// Responsibility is assigned by registered user IDs; names are display snapshots.
type EditInput struct {
	AssigneeID         string          `json:"assignee_id"`
	ReviewerID         string          `json:"reviewer_id"`
	ReleaseID          *string         `json:"release_id"`
	Title              string          `json:"title"`
	Description        string          `json:"description"`
	Priority           entity.Priority `json:"priority"`
	Product            string          `json:"product"`
	Reference          string          `json:"reference"`
	AcceptanceCriteria string          `json:"acceptance_criteria"`
	DueDate            *time.Time      `json:"due_date"`
	EventCodes         []string        `json:"event_codes"`
}

func (s *UseCase) Edit(ctx context.Context, id string, input EditInput) (entity.Demand, error) {
	if !identifier.IsUUID(id) {
		return entity.Demand{}, erring.ErrInvalidInput
	}

	for _, id := range []string{input.AssigneeID, input.ReviewerID} {
		if id != "" && !identifier.IsUUID(id) {
			return entity.Demand{}, erring.ErrInvalidInput.WithMessage("selecione um usuário cadastrado")
		}
	}

	var updated entity.Demand

	err := s.uow.Do(ctx, func(ctx context.Context, repos Repositories) error {
		item, err := repos.Demands().GetForUpdate(ctx, id)
		if err != nil {
			return erring.Wrap(err, "demand.callback")
		}

		if input.ReleaseID != nil {
			if *input.ReleaseID != "" && !identifier.IsUUID(*input.ReleaseID) {
				return erring.ErrInvalidInput.WithMessage("versão inválida")
			}

			item.ReleaseID = *input.ReleaseID
		}

		item.Title = strings.TrimSpace(input.Title)
		item.Description = strings.TrimSpace(input.Description)
		item.Priority = input.Priority
		item.AssigneeID = input.AssigneeID
		item.ReviewerID = input.ReviewerID
		item.Assignee = ""
		item.Reviewer = ""
		item.Product = strings.TrimSpace(input.Product)
		item.Reference = strings.TrimSpace(input.Reference)
		item.AcceptanceCriteria = strings.TrimSpace(input.AcceptanceCriteria)
		item.DueDate = input.DueDate
		item.EventCodes = input.EventCodes

		if item.EventCodes == nil {
			item.EventCodes = []string{}
		}

		item.UpdatedAt = s.now()
		if validationErr := item.Validate(); validationErr != nil {
			return erring.Wrap(validationErr, "demand.callback")
		}

		if updateErr := repos.Demands().UpdateDetails(ctx, item); updateErr != nil {
			return erring.Wrap(updateErr, "demand.callback")
		}

		updated, err = repos.Demands().GetForUpdate(ctx, id)

		return erring.Wrap(err, "demand.callback")
	})

	return updated, erring.Wrap(err, "demand.Edit")
}
