//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	workpackage "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/package"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres"
	packagerepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/package"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/testinfra"
)

func TestPostgresPackageRollbackAndList(t *testing.T) {
	t.Parallel()

	url := testinfra.Postgres(t)

	ctx := t.Context()

	store, err := postgres.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(store.Close)

	if err = store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	uow, err := postgres.NewUnitOfWork(store)
	if err != nil {
		t.Fatal(err)
	}

	var id string

	svc := workpackage.New(uow, packagerepo.NewPackageRepository(store.Pool), func() (string, error) { return "99999999-9999-4999-8999-999999999901", nil })

	result, err := svc.Create(ctx, workpackage.Input{Reference: "E2E-ROLLBACK", Title: "Pacote temporario de teste", Kind: entity.Improvement})
	if err != nil {
		t.Fatal(err)
	}

	id = result.ID

	t.Cleanup(func() {
		if _, cleanupErr := store.Pool.Exec(context.WithoutCancel(ctx), `DELETE FROM work_packages WHERE id=$1`, id); cleanupErr != nil {
			t.Error(cleanupErr)
		}
	})

	rollbackErr := errors.New("intentionally aborted")
	// A package change and any future item writes use the same transaction scope.
	err = uow.DoPackage(ctx, func(ctx context.Context, repos workpackage.Repositories) error {
		// Force an invalid second INSERT after first INSERT. Nothing must commit.
		demo := entity.WorkPackage{ID: "99999999-9999-4999-8999-999999999902", Title: "Rollback validation package", Kind: entity.Improvement, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if operationErr := repos.Packages().Insert(ctx, demo); operationErr != nil {
			return erring.Wrap(operationErr, "postgres_test.callback")
		}

		return rollbackErr
	})
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("want rollback error: %v", err)
	}

	if _, err = svc.Get(ctx, "99999999-9999-4999-8999-999999999902"); err == nil {
		t.Fatal("package persisted after rollback")
	}
}
