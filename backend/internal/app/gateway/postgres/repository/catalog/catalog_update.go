package catalogrepo

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/catalog"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *CatalogRepository) Update(ctx context.Context, id string, input catalog.Input) (entity.Workstream, error) {
	row, err := r.UpdateWorkstream(ctx, sqlc.UpdateWorkstreamParams{
		ID:     id,
		Name:   input.Name,
		Color:  input.Color,
		Active: input.Active,
	})
	if err != nil {
		return entity.Workstream{}, writeError(err)
	}

	return toWorkstream(row), nil
}
