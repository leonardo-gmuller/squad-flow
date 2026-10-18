//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/release"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres"
	demandrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/demand"
	releaserepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/release"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/testinfra"
)

func TestReleaseArchivePreservesHistoryAndPagesBoard(t *testing.T) {
	t.Parallel()

	url := testinfra.Postgres(t)

	ctx := t.Context()

	db, err := postgres.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(db.Close)

	if err = db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	unit, err := postgres.NewUnitOfWork(db)
	if err != nil {
		t.Fatal(err)
	}

	queries := demandrepo.NewDemandRepository(db.Pool)
	demands := demand.NewUseCase(unit, queries)
	versions := release.New(releaserepo.NewReleaseRepository(db.Pool))
	board := demand.NewBoardUseCase(queries)

	id, err := identifier.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	version, err := versions.Create(ctx, release.Input{Name: "V-" + id, Description: "Teste de versões"})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cleanup := context.WithoutCancel(ctx)
		if _, cleanupErr := db.Pool.Exec(cleanup, `DELETE FROM work_items WHERE release_id=$1 OR title LIKE $2`, version.ID, "Release test "+id+"%"); cleanupErr != nil {
			t.Error(cleanupErr)
		}

		if _, cleanupErr := db.Pool.Exec(cleanup, `DELETE FROM releases WHERE id=$1`, version.ID); cleanupErr != nil {
			t.Error(cleanupErr)
		}
	})

	if _, err = versions.Create(ctx, release.Input{Name: "  " + strings.ToLower(version.Name) + "  "}); !errors.Is(err, erring.ErrConflict) {
		t.Fatalf("duplicate release: %v", err)
	}

	if _, err = versions.Create(ctx, release.Input{Name: " "}); !errors.Is(err, erring.ErrInvalidInput) {
		t.Fatalf("blank release: %v", err)
	}

	completed := seedReleaseDemands(t, demands, version, id)
	assertReleasePages(t, board, versions, version)
	assertArchivedHistory(t, versions, demands, board, queries, version, completed, id)

	standalone, err := demands.Create(ctx, demand.CreateInput{Title: "Release test " + id + " standalone", Kind: entity.Corrective, Priority: "high"})
	if err != nil {
		t.Fatal(err)
	}

	none, err := board.Board(ctx, demand.BoardFilter{ReleaseID: "none", Search: standalone.Title})
	if err != nil {
		t.Fatal(err)
	}

	if none.Columns[0].Total != 1 {
		t.Fatal("unversioned demands missing")
	}

	if _, err = demands.Edit(ctx, standalone.ID, demand.EditInput{Title: standalone.Title, Priority: "high", ReleaseID: &version.ID}); !errors.Is(err, erring.ErrConflict) {
		t.Fatalf("reassigned to archived release: %v", err)
	}

	if err = versions.SetArchived(ctx, version.ID, false); err != nil {
		t.Fatal(err)
	}

	if _, err = demands.Edit(ctx, standalone.ID, demand.EditInput{Title: standalone.Title, Priority: "high", ReleaseID: &version.ID}); err != nil {
		t.Fatal(err)
	}

	restored, err := board.Board(ctx, demand.BoardFilter{ReleaseID: version.ID})
	if err != nil {
		t.Fatal(err)
	}

	if restored.Columns[0].Total != 23 || restored.Columns[4].Total != 1 {
		t.Fatalf("restore lost deliveries: %+v", restored)
	}

	empty := ""
	if _, err = demands.Edit(ctx, standalone.ID, demand.EditInput{Title: standalone.Title, Priority: "high", ReleaseID: &empty}); err != nil {
		t.Fatal(err)
	}

	detail, err := demands.Detail(ctx, standalone.ID)
	if err != nil || detail.Demand.ReleaseID != "" {
		t.Fatalf("clear version failed: %+v %v", detail, err)
	}
}

func assertReleasePages(t *testing.T, board *demand.BoardUseCase, versions *release.UseCase, version entity.Release) {
	t.Helper()
	ctx := t.Context()

	first, err := board.Board(ctx, demand.BoardFilter{ReleaseID: version.ID})
	if err != nil {
		t.Fatal(err)
	}

	if len(first.Columns) != 5 || first.Columns[0].Total != 22 || len(first.Columns[0].Items) != 20 || first.Columns[4].Total != 1 {
		t.Fatalf("wrong first page: %+v", first)
	}

	second, err := board.Board(ctx, demand.BoardFilter{ReleaseID: version.ID, Status: "backlog", Offset: 20, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}

	if len(second.Columns) != 1 || second.Columns[0].Total != 22 || len(second.Columns[0].Items) != 2 {
		t.Fatalf("wrong second page: %+v", second)
	}

	seen := map[string]bool{}
	for _, item := range first.Columns[0].Items {
		seen[item.ID] = true
	}

	for _, item := range second.Columns[0].Items {
		if seen[item.ID] {
			t.Fatal("duplicate demand between pages")
		}
	}

	versionsBefore, err := versions.List(ctx)
	if err != nil {
		t.Fatal(err)
	}

	for _, item := range versionsBefore {
		if item.ID == version.ID && (item.Total != 23 || item.Done != 1) {
			t.Fatalf("wrong release counters: %+v", item)
		}
	}
}

func assertArchivedHistory(t *testing.T, versions *release.UseCase, demands *demand.UseCase, board *demand.BoardUseCase, queries *demandrepo.DemandRepository, version entity.Release, completed entity.Demand, id string) {
	t.Helper()
	ctx := t.Context()

	var err error

	if err = versions.SetArchived(ctx, version.ID, true); err != nil {
		t.Fatal(err)
	}

	active, err := board.Board(ctx, demand.BoardFilter{ReleaseID: version.ID})
	if err != nil {
		t.Fatal(err)
	}

	for _, column := range active.Columns {
		if column.Total != 0 {
			t.Fatal("archived release visible in active board")
		}
	}

	archived, err := board.Board(ctx, demand.BoardFilter{ReleaseID: version.ID, Archive: "archived"})
	if err != nil {
		t.Fatal(err)
	}

	if archived.Columns[0].Total != 22 || archived.Columns[4].Total != 1 {
		t.Fatalf("archive lost cards: %+v", archived)
	}

	if _, err = demands.Create(ctx, demand.CreateInput{Title: "Release test " + id + " forbidden", ReleaseID: version.ID, Kind: entity.Improvement, Priority: "medium"}); !errors.Is(err, erring.ErrConflict) {
		t.Fatalf("assigned archived release: %v", err)
	}
	// Omitting release_id in legacy edit clients preserves the existing version and QA state.
	updated, err := demands.Edit(ctx, completed.ID, demand.EditInput{Title: completed.Title, Priority: "medium"})
	if err != nil {
		t.Fatal(err)
	}

	if updated.ReleaseID != version.ID || updated.ReleaseName != version.Name || updated.Status != entity.Done {
		t.Fatalf("edit lost release or QA state: %+v", updated)
	}

	detail, err := demands.Detail(ctx, completed.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(detail.Tests) != 1 || len(detail.Transitions) != 3 || detail.Demand.ReleaseName != version.Name {
		t.Fatalf("history not preserved: %+v", detail)
	}

	historical, err := queries.AllForMetrics(ctx)
	if err != nil {
		t.Fatal(err)
	}

	found := false

	for _, item := range historical {
		if item.ID == completed.ID {
			found = true
		}
	}

	if !found {
		t.Fatal("archived delivery excluded from historical metrics")
	}
}

func seedReleaseDemands(t *testing.T, demands *demand.UseCase, version entity.Release, id string) entity.Demand {
	t.Helper()
	ctx := t.Context()

	var err error

	var completed entity.Demand

	for index := range 23 {
		item, createErr := demands.Create(ctx, demand.CreateInput{Title: fmt.Sprintf("Release test %s %02d", id, index), ReleaseID: version.ID, Kind: entity.Improvement, Priority: "medium"})
		if createErr != nil {
			t.Fatal(createErr)
		}

		if index == 0 {
			completed = item
		}
	}

	for _, status := range []entity.Status{entity.InProgress, entity.InQA} {
		if _, err = demands.Move(ctx, completed.ID, demand.MoveInput{Status: status, Actor: "QA"}); err != nil {
			t.Fatal(err)
		}
	}

	if _, err = demands.AddTest(ctx, completed.ID, demand.AddTestInput{Title: "Homologação da versão", Result: "passed", Author: "QA"}); err != nil {
		t.Fatal(err)
	}

	if _, err = demands.Move(ctx, completed.ID, demand.MoveInput{Status: entity.Done, Actor: "QA"}); err != nil {
		t.Fatal(err)
	}

	return completed
}
