package testcaserepo

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *TestCaseRepository) CurrentStats(ctx context.Context, id string, cycle int) (int, int, error) {
	row, err := r.GetQACycleStats(ctx, sqlc.GetQACycleStatsParams{DemandID: id, Cycle: cycle})

	return row.Passed, row.Failed, erring.Wrap(err, "testcaserepo.CurrentStats")
}
