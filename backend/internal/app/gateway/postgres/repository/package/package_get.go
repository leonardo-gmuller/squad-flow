package packagerepo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *PackageRepository) List(ctx context.Context) ([]entity.WorkPackage, error) {
	rows, err := r.ListPackages(ctx)
	if err != nil {
		return nil, erring.Wrap(err, "packagerepo.List")
	}

	items := make([]entity.WorkPackage, 0, len(rows))
	for _, row := range rows {
		items = append(items, toPackage(sqlc.GetPackageRow(row)))
	}

	return items, nil
}

func (r *PackageRepository) Get(ctx context.Context, id string) (entity.WorkPackage, error) {
	row, err := r.GetPackage(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.WorkPackage{}, erring.ErrNotFound
	}

	if err != nil {
		return entity.WorkPackage{}, erring.Wrap(err, "packagerepo.Get")
	}

	return toPackage(row), nil
}
