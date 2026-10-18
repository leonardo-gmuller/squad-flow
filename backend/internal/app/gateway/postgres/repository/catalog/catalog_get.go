package catalogrepo

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

func (r *CatalogRepository) List(ctx context.Context) ([]entity.Workstream, error) {
	rows, err := r.ListWorkstreams(ctx)
	if err != nil {
		return nil, erring.Wrap(err, "catalogrepo.List")
	}

	items := make([]entity.Workstream, 0, len(rows))
	for _, row := range rows {
		items = append(items, toWorkstream(row))
	}

	return items, nil
}

func (r *CatalogRepository) Members(ctx context.Context) ([]entity.Member, error) {
	rows, err := r.ListMembers(ctx)
	if err != nil {
		return nil, erring.Wrap(err, "catalogrepo.Members")
	}

	items := make([]entity.Member, 0, len(rows))
	for _, row := range rows {
		items = append(items, toMember(row))
	}

	return items, nil
}
