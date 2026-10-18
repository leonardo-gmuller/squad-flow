package transitionrepo

import (
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func createTransitionParams(item entity.Transition) sqlc.CreateTransitionParams {
	return sqlc.CreateTransitionParams{
		ID:         item.ID,
		DemandID:   item.DemandID,
		FromStatus: string(item.From),
		ToStatus:   string(item.To),
		Actor:      item.Actor,
		Reason:     item.Reason,
		CreatedAt:  item.CreatedAt,
		ActorID:    item.ActorID,
	}
}
