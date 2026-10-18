package demandrepo

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	notificationrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/notification"
)

func (r *DemandRepository) Update(ctx context.Context, item entity.Demand) error {
	before, err := r.GetForUpdate(ctx, item.ID)
	if err != nil {
		return err
	}

	affected, err := r.UpdateDemandStatus(ctx, updateDemandStatusParams(item))
	if err != nil {
		return erring.Wrap(err, "demandrepo.Update")
	}

	if affected != 1 {
		return erring.ErrNotFound
	}

	return erring.Wrap(notificationrepo.StatusChanged(ctx, r.Queries, before, item), "demandrepo.Update.notifications")
}

func (r *DemandRepository) UpdateDetails(ctx context.Context, item entity.Demand) error {
	before, err := r.GetForUpdate(ctx, item.ID)
	if err != nil {
		return err
	}

	if err = r.validateAssignments(ctx, &item, true); err != nil {
		return err
	}

	if _, err = r.validateRelease(ctx, item.ReleaseID, item.ID); err != nil {
		return err
	}

	affected, err := r.UpdateDemandDetails(ctx, updateDemandDetailsParams(item))
	if err != nil {
		return erring.Wrap(err, "demandrepo.UpdateDetails")
	}

	if affected != 1 {
		return erring.ErrNotFound
	}

	return erring.Wrap(notificationrepo.Assignments(ctx, r.Queries, before, item), "demandrepo.UpdateDetails.notifications")
}
