package demand

import (
	"context"
	"strings"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

type CreateInput struct {
	AssigneeID         string          `json:"assignee_id"`
	ReviewerID         string          `json:"reviewer_id"`
	Title              string          `json:"title"`
	Description        string          `json:"description"`
	Kind               entity.Kind     `json:"kind"`
	Priority           entity.Priority `json:"priority"`
	Reference          string          `json:"reference"`
	PackageID          string          `json:"package_id"`
	ReleaseID          string          `json:"release_id"`
	Product            string          `json:"product"`
	AcceptanceCriteria string          `json:"acceptance_criteria"`
	DueDate            *time.Time      `json:"due_date"`
	EventCodes         []string        `json:"event_codes"`
}

func (s *UseCase) Create(ctx context.Context, input CreateInput) (entity.Demand, error) {
	id, err := identifier.NewUUID()
	if err != nil {
		return entity.Demand{}, erring.Wrap(err, "demand.Create.uuid")
	}

	input.PackageID = strings.TrimSpace(input.PackageID)
	if input.PackageID != "" && !identifier.IsUUID(input.PackageID) {
		return entity.Demand{}, erring.ErrInvalidInput.WithMessage("OS inválida")
	}

	if input.ReleaseID != "" && !identifier.IsUUID(input.ReleaseID) {
		return entity.Demand{}, erring.ErrInvalidInput.WithMessage("versão inválida")
	}

	for _, id := range []string{input.AssigneeID, input.ReviewerID} {
		if id != "" && !identifier.IsUUID(id) {
			return entity.Demand{}, erring.ErrInvalidInput.WithMessage("selecione um usuário cadastrado")
		}
	}

	now := s.now()

	if input.EventCodes == nil {
		input.EventCodes = []string{}
	}

	item := entity.Demand{
		ID: id, Title: strings.TrimSpace(input.Title), Description: strings.TrimSpace(input.Description),
		AssigneeID: input.AssigneeID, ReviewerID: input.ReviewerID, Kind: input.Kind, Priority: input.Priority, Status: entity.Backlog,
		Reference: strings.TrimSpace(input.Reference),
		ReleaseID: input.ReleaseID, PackageID: strings.TrimSpace(input.PackageID), Product: strings.TrimSpace(input.Product),
		AcceptanceCriteria: strings.TrimSpace(input.AcceptanceCriteria), DueDate: input.DueDate,
		EventCodes: input.EventCodes, CreatedAt: now, UpdatedAt: now,
	}
	if createErr := item.Validate(); createErr != nil {
		return entity.Demand{}, erring.Wrap(createErr, "demand.Create")
	}

	err = s.uow.Do(ctx, func(ctx context.Context, repos Repositories) error {
		return repos.Demands().Insert(ctx, &item)
	})

	return item, erring.Wrap(err, "demand.Create")
}
