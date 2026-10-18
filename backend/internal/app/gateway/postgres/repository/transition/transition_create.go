package transitionrepo

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

func (r *TransitionRepository) Insert(ctx context.Context, item entity.Transition) error {
	return erring.Wrap(r.CreateTransition(ctx, createTransitionParams(item)), "transitionrepo.Insert")
}
