package demand

import (
	"context"
	"fmt"
	"strings"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

type MoveInput struct {
	ActorID string        `json:"-"`
	Status  entity.Status `json:"status"`
	Actor   string        `json:"actor"`
	Reason  string        `json:"reason"`
}

func (s *UseCase) Move(ctx context.Context, id string, input MoveInput) (entity.Demand, error) {
	var result entity.Demand
	if err := input.validate(id); err != nil {
		return result, err
	}

	err := s.uow.Do(ctx, func(ctx context.Context, repos Repositories) error {
		item, err := repos.Demands().GetForUpdate(ctx, id)
		if err != nil {
			return erring.Wrap(err, "demand.callback")
		}

		if !entity.CanTransition(item.Status, input.Status) {
			return erring.ErrInvalidTransition.WithMessage(fmt.Sprintf("Movimentação de %s para %s não permitida.", item.Status, input.Status))
		}

		if input.Status == entity.Done {
			passed, failed, operationErr := repos.Tests().CurrentStats(ctx, id, item.QACycle)
			if operationErr != nil {
				return erring.Wrap(operationErr, "demand.callback")
			}

			if passed == 0 || failed > 0 {
				return erring.ErrConflict.WithMessage("A conclusão exige ao menos um teste aprovado e nenhum reprovado no ciclo atual.")
			}
		}

		from := item.Status
		item.Status = input.Status
		item.UpdatedAt = s.now()

		if input.Status == entity.InQA {
			item.QACycle++
		}

		if input.Status == entity.ChangesRequested {
			item.ReturnCount++
		}

		if input.Status == entity.Done {
			completed := item.UpdatedAt
			item.CompletedAt = &completed
		}

		if callbackErr2 := repos.Demands().Update(ctx, item); callbackErr2 != nil {
			return erring.Wrap(callbackErr2, "demand.callback")
		}

		transitionID, err := identifier.NewUUID()
		if err != nil {
			return erring.Wrap(err, "demand.Move.uuid")
		}

		if err := repos.Transitions().Insert(ctx, entity.Transition{ID: transitionID, DemandID: id, From: from, To: input.Status, ActorID: input.ActorID, Actor: strings.TrimSpace(input.Actor), Reason: strings.TrimSpace(input.Reason), CreatedAt: item.UpdatedAt}); err != nil {
			return erring.Wrap(err, "demand.callback")
		}

		result = item

		return nil
	})

	return result, erring.Wrap(err, "demand.Move")
}

func (input MoveInput) validate(id string) error {
	if !identifier.IsUUID(id) {
		return erring.ErrInvalidInput
	}

	if !entity.IsStatus(input.Status) {
		return erring.ErrInvalidInput.WithMessage("Etapa inválida.")
	}

	if len(input.Actor) > 120 || len(input.Reason) > 2000 {
		return erring.ErrInvalidInput.WithMessage("Um dos campos excede o tamanho permitido.")
	}

	if input.Status == entity.ChangesRequested && strings.TrimSpace(input.Reason) == "" {
		return erring.ErrInvalidInput.WithMessage("Informe o motivo da movimentação.")
	}

	return nil
}
