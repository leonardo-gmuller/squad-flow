package demandrepo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	notificationrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/notification"
)

func (r *DemandRepository) Insert(ctx context.Context, item *entity.Demand) error {
	if err := r.validateAssignments(ctx, item, false); err != nil {
		return err
	}

	name, err := r.validateRelease(ctx, item.ReleaseID, item.ID)
	if err != nil {
		return err
	}

	item.ReleaseName = name

	number, err := r.CreateDemand(ctx, createDemandParams(*item))
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && (pe.Code == "23503" || pe.Code == "22P02") {
			return erring.ErrInvalidInput.WithMessage("OS inválida.")
		}

		return erring.Wrap(err, "demandrepo.Insert")
	}

	item.Number = number

	return erring.Wrap(notificationrepo.Assignments(ctx, r.Queries, entity.Demand{}, *item), "demandrepo.Insert.notifications")
}
