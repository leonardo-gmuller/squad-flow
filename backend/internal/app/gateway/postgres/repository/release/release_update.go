package releaserepo

import (
	"context"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/release"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *ReleaseRepository) Update(ctx context.Context, id string, input release.Input, now time.Time) error {
	affected, err := r.UpdateRelease(ctx, sqlc.UpdateReleaseParams{
		ID:          id,
		Name:        input.Name,
		Description: input.Description,
		ReleaseDate: input.ReleaseDate,
		UpdatedAt:   now,
	})
	if err != nil {
		return writeError(err)
	}

	if affected == 0 {
		return erring.ErrNotFound
	}

	return nil
}

func (r *ReleaseRepository) SetArchived(ctx context.Context, id string, archived bool, now time.Time) error {
	affected, err := r.SetReleaseArchived(ctx, sqlc.SetReleaseArchivedParams{
		ID:        id,
		Archived:  archived,
		UpdatedAt: now,
	})
	if err != nil {
		return writeError(err)
	}

	if affected == 0 {
		return erring.ErrNotFound
	}

	return nil
}
