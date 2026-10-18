//go:build integration

package postgres_test

import (
	"errors"
	"testing"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	workpackage "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/package"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres"
	demandrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/demand"
	packagerepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/package"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/testinfra"
)

func TestPostgresQACycleWithRejectionAndFinalApproval(t *testing.T) {
	t.Parallel()

	store, err := postgres.New(t.Context(), testinfra.Postgres(t))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(store.Close)

	if err = store.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}

	uow, err := postgres.NewUnitOfWork(store)
	if err != nil {
		t.Fatal(err)
	}

	service := demand.NewUseCase(uow, demandrepo.NewDemandRepository(store.Pool))

	item, err := service.Create(t.Context(), demand.CreateInput{Title: "Nova regra do evento", Kind: entity.Improvement, Priority: "high", EventCodes: []string{"S-2200"}})
	if err != nil {
		t.Fatal(err)
	}

	move := func(status entity.Status) {
		t.Helper()

		if _, moveError := service.Move(t.Context(), item.ID, demand.MoveInput{Status: status, Actor: "Developer"}); moveError != nil {
			t.Fatal(moveError)
		}
	}
	move(entity.InProgress)
	move(entity.InQA)

	if _, err = service.Move(t.Context(), item.ID, demand.MoveInput{Status: entity.Done}); !errors.Is(err, erring.ErrConflict) {
		t.Fatalf("completion without tests: %v", err)
	}

	failed, err := service.AddTest(t.Context(), item.ID, demand.AddTestInput{Title: "Validação incorreta", Result: "failed", Actual: "Campo obrigatório ausente", Author: "QA"})
	if err != nil {
		t.Fatal(err)
	}

	returned, err := service.Detail(t.Context(), item.ID)
	if err != nil {
		t.Fatal(err)
	}

	if failed.Cycle != 1 || returned.Demand.Status != entity.ChangesRequested || returned.Demand.ReturnCount != 1 {
		t.Fatalf("QA failure did not return the demand: %+v", returned)
	}

	move(entity.InProgress)
	move(entity.InQA)

	if _, err = service.Move(t.Context(), item.ID, demand.MoveInput{Status: entity.Done}); !errors.Is(err, erring.ErrConflict) {
		t.Fatalf("old tests were accepted for a new cycle: %v", err)
	}

	passed, err := service.AddTest(t.Context(), item.ID, demand.AddTestInput{Title: "Validação corrigida", Result: "passed", Author: "QA"})
	if err != nil {
		t.Fatal(err)
	}

	if passed.Cycle != 2 {
		t.Fatalf("incorrect QA cycle: %d", passed.Cycle)
	}

	move(entity.Done)

	detail, err := service.Detail(t.Context(), item.ID)
	if err != nil {
		t.Fatal(err)
	}

	if detail.Demand.CompletedAt == nil || detail.Demand.Status != entity.Done || detail.Demand.ReturnCount != 1 || len(detail.Tests) != 2 || len(detail.Transitions) != 6 {
		t.Fatalf("workflow history not persisted: %+v", detail)
	}

	metrics, err := service.Metrics(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if metrics.Delivered != 1 || metrics.FirstPassPercent != 0 {
		t.Fatalf("incorrect persisted QA metrics: %+v", metrics)
	}
}

func TestPostgresRollsBackPackageWhenFirstItemFails(t *testing.T) {
	t.Parallel()

	store, err := postgres.New(t.Context(), testinfra.Postgres(t))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(store.Close)

	if err = store.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}

	uow, err := postgres.NewUnitOfWork(store)
	if err != nil {
		t.Fatal(err)
	}

	packageID := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	next := 0
	service := workpackage.New(uow, packagerepo.NewPackageRepository(store.Pool), func() (string, error) {
		next++
		if next == 1 {
			return packageID, nil
		}

		return "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", nil
	})
	// The first package INSERT succeeds; a missing user makes the demand INSERT fail.
	result, err := service.Create(t.Context(), workpackage.Input{Title: "Pacote com responsável inexistente", Kind: entity.Improvement, CreateFirstItem: true, AssigneeID: "55555555-5555-4555-8555-555555555599"})
	if err == nil || result.ID != "" {
		t.Fatalf("invalid first item was accepted: %+v error=%v", result, err)
	}

	if _, err = service.Get(t.Context(), packageID); !errors.Is(err, erring.ErrNotFound) {
		t.Fatalf("partial package survived a failed transaction: %v", err)
	}
}
