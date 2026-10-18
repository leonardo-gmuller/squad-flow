package releaserepo

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
)

func (r *ReleaseRepository) Insert(ctx context.Context, item entity.Release) error {
	return writeError(r.CreateRelease(ctx, createReleaseParams(item)))
}
