package catalogrepo

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *CatalogRepository) Insert(ctx context.Context, item entity.Workstream) (entity.Workstream, error) {
	row, err := r.CreateWorkstream(ctx, sqlc.CreateWorkstreamParams{
		ID:     item.ID,
		Code:   item.Code,
		Name:   item.Name,
		Color:  item.Color,
		Active: item.Active,
	})
	if err != nil {
		return entity.Workstream{}, writeError(err)
	}

	return toWorkstream(row), nil
}
