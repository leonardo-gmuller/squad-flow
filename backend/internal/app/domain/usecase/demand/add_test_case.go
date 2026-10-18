package demand

import (
	"context"
	"strings"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

type AddTestInput struct {
	AuthorID string `json:"-"`
	Title    string `json:"title"`
	Steps    string `json:"steps"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Result   string `json:"result"`
	Author   string `json:"author"`
}

func (s *UseCase) AddTest(ctx context.Context, id string, input AddTestInput) (entity.TestCase, error) {
	if !identifier.IsUUID(id) {
		return entity.TestCase{}, erring.ErrInvalidInput
	}

	testID, err := identifier.NewUUID()
	if err != nil {
		return entity.TestCase{}, erring.Wrap(err, "demand.AddTest.uuid")
	}

	tc := entity.TestCase{ID: testID, DemandID: id, Title: strings.TrimSpace(input.Title), Steps: input.Steps, Expected: input.Expected, Actual: input.Actual, Result: input.Result, AuthorID: input.AuthorID, Author: strings.TrimSpace(input.Author), CreatedAt: s.now()}
	if addTestErr := tc.Validate(); addTestErr != nil {
		return entity.TestCase{}, erring.Wrap(addTestErr, "demand.AddTest")
	}

	err = s.uow.Do(ctx, func(ctx context.Context, repos Repositories) error {
		item, operationErr := repos.Demands().GetForUpdate(ctx, id)
		if operationErr != nil {
			return erring.Wrap(operationErr, "demand.callback")
		}

		if item.Status != entity.InQA {
			return erring.ErrConflict.WithMessage("Testes só podem ser registrados durante o QA.")
		}

		tc.Cycle = item.QACycle
		if insertErr := repos.Tests().Insert(ctx, tc); insertErr != nil {
			return erring.Wrap(insertErr, "demand.callback")
		}

		if tc.Result == "failed" {
			from := item.Status
			item.Status = entity.ChangesRequested
			item.ReturnCount++

			item.UpdatedAt = tc.CreatedAt
			if updateErr := repos.Demands().Update(ctx, item); updateErr != nil {
				return erring.Wrap(updateErr, "demand.callback")
			}

			trID, idErr := identifier.NewUUID()
			if idErr != nil {
				return erring.Wrap(idErr, "demand.AddTest.transitionUUID")
			}

			if transitionErr := repos.Transitions().Insert(ctx, entity.Transition{ID: trID, DemandID: id, From: from, To: item.Status, ActorID: tc.AuthorID, Actor: tc.Author, Reason: "QA: " + tc.Title + " — " + tc.Actual, CreatedAt: tc.CreatedAt}); transitionErr != nil {
				return erring.Wrap(transitionErr, "demand.callback")
			}
		}

		return nil
	})

	return tc, erring.Wrap(err, "demand.AddTest")
}
