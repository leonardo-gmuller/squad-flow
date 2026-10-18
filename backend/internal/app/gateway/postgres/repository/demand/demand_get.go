package demandrepo

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *DemandRepository) GetForUpdate(ctx context.Context, id string) (entity.Demand, error) {
	row, err := r.GetDemandForUpdate(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Demand{}, erring.ErrNotFound
	}

	if err != nil {
		return entity.Demand{}, erring.Wrap(err, "demandrepo.GetForUpdate")
	}

	return toDemand(sqlc.GetDemandRow(row)), nil
}

func (r *DemandRepository) List(ctx context.Context, kind, status, search string) ([]entity.Demand, error) {
	if kind != "" && !entity.IsKind(entity.Kind(kind)) {
		return nil, erring.ErrInvalidInput.WithMessage("Esteira inválida.")
	}

	if status != "" && !entity.IsStatus(entity.Status(status)) {
		return nil, erring.ErrInvalidInput.WithMessage("Etapa inválida.")
	}

	rows, err := r.ListDemands(ctx, sqlc.ListDemandsParams{
		Kind:   kind,
		Status: status,
		Search: strings.TrimSpace(search),
	})
	if err != nil {
		return nil, erring.Wrap(err, "demandrepo.List")
	}

	items := make([]entity.Demand, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDemand(sqlc.GetDemandRow(row)))
	}

	return items, nil
}

func (r *DemandRepository) AllForMetrics(ctx context.Context) ([]entity.Demand, error) {
	return r.List(ctx, "", "", "")
}

func (r *DemandRepository) Detail(ctx context.Context, id string) (demand.Detail, error) {
	var detail demand.Detail

	row, err := r.GetDemand(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return detail, erring.ErrNotFound
	}

	if err != nil {
		return detail, erring.Wrap(err, "demandrepo.Detail")
	}

	detail.Demand = toDemand(row)

	tests, err := r.ListDemandTests(ctx, id)
	if err != nil {
		return detail, erring.Wrap(err, "demandrepo.Detail.tests")
	}

	detail.Tests = make([]entity.TestCase, 0, len(tests))
	for _, test := range tests {
		detail.Tests = append(detail.Tests, toTestCase(test))
	}

	transitions, err := r.ListDemandTransitions(ctx, id)
	if err != nil {
		return detail, erring.Wrap(err, "demandrepo.Detail.transitions")
	}

	detail.Transitions = make([]entity.Transition, 0, len(transitions))
	for _, transition := range transitions {
		detail.Transitions = append(detail.Transitions, toTransition(transition))
	}

	return detail, nil
}
