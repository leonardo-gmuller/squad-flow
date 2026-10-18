package workpackage

import (
	"context"
	"strings"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

type Input struct {
	AssigneeID      string      `json:"assignee_id"`
	ReviewerID      string      `json:"reviewer_id"`
	ReleaseID       string      `json:"release_id"`
	Reference       string      `json:"reference"`
	Title           string      `json:"title"`
	Description     string      `json:"description"`
	Initiative      string      `json:"initiative"`
	Kind            entity.Kind `json:"kind"`
	CreateFirstItem bool        `json:"create_first_item"`
	Product         string      `json:"product"`
}
type PackageUseCase struct {
	uow     UnitOfWork
	queries Repository
	now     func() time.Time
	newID   func() (string, error)
}

func New(uow UnitOfWork, queries Repository, newID func() (string, error)) *PackageUseCase {
	return &PackageUseCase{uow: uow, queries: queries, newID: newID, now: func() time.Time { return time.Now().UTC() }}
}

func (s *PackageUseCase) List(ctx context.Context) ([]entity.WorkPackage, error) {
	result, queryErr := s.queries.List(ctx)

	return result, erring.Wrap(queryErr, "workpackage.List")
}

func (s *PackageUseCase) Get(ctx context.Context, id string) (entity.WorkPackage, error) {
	if !identifier.IsUUID(id) {
		return entity.WorkPackage{}, erring.ErrInvalidInput
	}

	result, queryErr := s.queries.Get(ctx, id)

	return result, erring.Wrap(queryErr, "workpackage.Get")
}

func (s *PackageUseCase) Create(ctx context.Context, input Input) (entity.WorkPackage, error) {
	id, err := s.newID()
	if err != nil {
		return entity.WorkPackage{}, err
	}

	now := s.now()

	workPackage := entity.WorkPackage{ID: id, Reference: strings.TrimSpace(input.Reference), Title: strings.TrimSpace(input.Title), Description: strings.TrimSpace(input.Description), Initiative: strings.TrimSpace(input.Initiative), Kind: input.Kind, CreatedAt: now, UpdatedAt: now}
	if err = workPackage.Validate(); err != nil {
		return entity.WorkPackage{}, erring.Wrap(err, "workpackage.Create")
	}

	if (input.AssigneeID != "" && !identifier.IsUUID(input.AssigneeID)) || (input.ReviewerID != "" && !identifier.IsUUID(input.ReviewerID)) || len(input.Product) > 120 {
		return workPackage, erring.ErrInvalidInput.WithMessage("responsável inválido ou produto muito longo")
	}

	if input.ReleaseID != "" && !identifier.IsUUID(input.ReleaseID) {
		return entity.WorkPackage{}, erring.ErrInvalidInput.WithMessage("versão inválida")
	}

	var item entity.Demand

	if input.CreateFirstItem {
		itemID, createErr := s.newID()
		if createErr != nil {
			return workPackage, createErr
		}

		item = entity.Demand{AssigneeID: input.AssigneeID, ReviewerID: input.ReviewerID, ID: itemID, Title: workPackage.Title, Description: workPackage.Description, PackageID: workPackage.ID, ReleaseID: input.ReleaseID, Kind: workPackage.Kind, Priority: "medium", Status: entity.Backlog, Product: strings.TrimSpace(input.Product), Reference: workPackage.Reference, EventCodes: []string{}, CreatedAt: now, UpdatedAt: now}
		if CreateErr2 := item.Validate(); CreateErr2 != nil {
			return workPackage, erring.Wrap(CreateErr2, "workpackage.Create")
		}
	}

	err = s.uow.DoPackage(ctx, func(ctx context.Context, repos Repositories) error {
		if operationErr := repos.Packages().Insert(ctx, workPackage); operationErr != nil {
			return erring.Wrap(operationErr, "workpackage.callback")
		}

		if input.CreateFirstItem {
			return repos.Demands().Insert(ctx, &item)
		}

		return nil
	})
	if err != nil {
		return entity.WorkPackage{}, erring.Wrap(err, "workpackage.Create")
	}

	if input.CreateFirstItem {
		workPackage.Total = 1
	}

	return workPackage, nil
}
