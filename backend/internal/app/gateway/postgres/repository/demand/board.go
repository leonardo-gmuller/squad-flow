package demandrepo

import (
	"context"
	"encoding/json"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

// All column counts and pages come from the same database snapshot.
func (r *DemandRepository) Board(ctx context.Context, filter demand.BoardFilter) (demand.Board, error) {
	rows, err := r.GetBoardColumns(ctx, sqlc.GetBoardColumnsParams{
		ReleaseID:  filter.ReleaseID,
		Archive:    filter.Archive,
		Kind:       filter.Kind,
		Owner:      filter.Owner,
		Search:     filter.Search,
		Status:     filter.Status,
		PageOffset: filter.Offset,
		PageLimit:  filter.Limit,
	})
	if err != nil {
		return demand.Board{}, erring.Wrap(err, "demandrepo.Board")
	}

	byStatus := make(map[entity.Status]demand.BoardColumn, len(rows))

	for _, row := range rows {
		column := demand.BoardColumn{Status: entity.Status(row.Status), Total: row.Total}
		if err = json.Unmarshal(row.Items, &column.Items); err != nil {
			return demand.Board{}, erring.Wrap(err, "demandrepo.Board")
		}

		byStatus[column.Status] = column
	}

	out := demand.Board{Columns: []demand.BoardColumn{}, Owners: []string{}}

	for _, status := range []entity.Status{
		entity.Backlog,
		entity.InProgress,
		entity.InQA,
		entity.ChangesRequested,
		entity.Done,
	} {
		if filter.Status != "" && string(status) != filter.Status {
			continue
		}

		column, exists := byStatus[status]
		if !exists {
			column = demand.BoardColumn{Status: status, Items: []entity.Demand{}}
		}

		out.Columns = append(out.Columns, column)
	}

	out.Owners, err = r.ListBoardOwners(ctx)

	return out, erring.Wrap(err, "demandrepo.Board.owners")
}
