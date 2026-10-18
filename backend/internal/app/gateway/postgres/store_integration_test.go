//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres"
	demandrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/demand"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/testinfra"
)

func TestPostgresUnitOfWorkRollback(t *testing.T) {
	t.Parallel()

	url := testinfra.Postgres(t)

	ctx := t.Context()

	store, err := postgres.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}

	defer store.Close()

	if testPostgresUnitOfWorkRollbackErr := store.Migrate(ctx); testPostgresUnitOfWorkRollbackErr != nil {
		t.Fatal(testPostgresUnitOfWorkRollbackErr)
	}

	uow, err := postgres.NewUnitOfWork(store)
	if err != nil {
		t.Fatal(err)
	}

	service := demand.NewUseCase(uow, demandrepo.NewDemandRepository(store.Pool))

	createdDemand, err := service.Create(ctx, demand.CreateInput{Title: "Teste de transação com rollback", Kind: entity.Improvement, Priority: "medium"})
	if err != nil {
		t.Fatal(err)
	}

	rolledBack := errors.New("rollback requested")
	err = uow.Do(ctx, func(ctx context.Context, repos demand.Repositories) error {
		item, operationErr := repos.Demands().GetForUpdate(ctx, createdDemand.ID)
		if operationErr != nil {
			return erring.Wrap(operationErr, "postgres_test.callback")
		}

		item.Status = entity.InProgress
		if updateErr := repos.Demands().Update(ctx, item); updateErr != nil {
			return erring.Wrap(updateErr, "postgres_test.callback")
		}

		return rolledBack
	})

	if !errors.Is(err, rolledBack) {
		t.Fatalf("want rollback error: %v", err)
	}

	detail, err := service.Detail(ctx, createdDemand.ID)
	if err != nil {
		t.Fatal(err)
	}

	if detail.Demand.Status != entity.Backlog {
		t.Fatalf("status persisted after rollback: %s", detail.Demand.Status)
	}
}
