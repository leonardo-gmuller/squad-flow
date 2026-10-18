package packagerepo

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	catalogrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/catalog"
)

func (r *PackageRepository) Insert(ctx context.Context, item entity.WorkPackage) error {
	if err := catalogrepo.ValidateKind(ctx, r.Queries, item.Kind, false); err != nil {
		return erring.Wrap(err, "packagerepo.Insert.kind")
	}

	return erring.Wrap(r.CreatePackage(ctx, createPackageParams(item)), "packagerepo.Insert")
}
