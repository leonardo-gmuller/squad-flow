package demand

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

func (s *UseCase) List(ctx context.Context, kind, status, q string) ([]entity.Demand, error) {
	result, queryErr := s.queries.List(ctx, kind, status, q)

	return result, erring.Wrap(queryErr, "demand.List")
}

func (s *UseCase) Detail(ctx context.Context, id string) (Detail, error) {
	if !identifier.IsUUID(id) {
		return Detail{}, erring.ErrInvalidInput
	}

	result, queryErr := s.queries.Detail(ctx, id)

	return result, erring.Wrap(queryErr, "demand.Detail")
}
