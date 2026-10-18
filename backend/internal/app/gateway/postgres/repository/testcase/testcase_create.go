package testcaserepo

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	notificationrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/notification"
)

func (r *TestCaseRepository) Insert(ctx context.Context, item entity.TestCase) error {
	if err := r.CreateTestCase(ctx, createTestCaseParams(item)); err != nil {
		return erring.Wrap(err, "testcaserepo.Insert")
	}

	if item.Result == "passed" {
		return erring.Wrap(notificationrepo.QAPassed(ctx, r.Queries, item.DemandID, item.Title), "testcaserepo.Insert.notifications")
	}

	return nil
}
