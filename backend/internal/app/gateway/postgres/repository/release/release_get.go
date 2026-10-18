package releaserepo

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

func (r *ReleaseRepository) List(ctx context.Context) ([]entity.Release, error) {
	rows, err := r.ListReleases(ctx)
	if err != nil {
		return nil, erring.Wrap(err, "releaserepo.List")
	}

	items := make([]entity.Release, 0, len(rows))
	for _, row := range rows {
		items = append(items, toRelease(row))
	}

	return items, nil
}
