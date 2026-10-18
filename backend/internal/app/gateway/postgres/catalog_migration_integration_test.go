//go:build integration

package postgres_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/testinfra"
)

func TestCatalogMigrationPreservesLegacyAndMatchesOnlyUniqueUsers(t *testing.T) {
	t.Parallel()

	databaseURL := testinfra.Postgres(t)

	ctx := t.Context()

	base, err := postgres.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	defer base.Close()

	id, _ := identifier.NewUUID()

	schema := "catalog_migration_" + strings.ReplaceAll(id, "-", "")
	if _, err = base.Pool.Exec(ctx, `CREATE SCHEMA `+pgx.Identifier{schema}.Sanitize()); err != nil {
		t.Fatal(err)
	}

	defer func() {
		_, _ = base.Pool.Exec(context.WithoutCancel(ctx), `DROP SCHEMA `+pgx.Identifier{schema}.Sanitize()+` CASCADE`)
	}()

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	cfg.ConnConfig.RuntimeParams["search_path"] = schema

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}

	defer pool.Close()
	db := &postgres.Client{Pool: pool}

	if _, err = pool.Exec(ctx, `CREATE TABLE schema_migrations(version text PRIMARY KEY,applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatal(err)
	}

	for _, file := range []string{"001_init.sql", "002_work_packages.sql", "003_auth_reports.sql", "004_releases.sql", "005_password_reset.sql"} {
		raw, readErr := os.ReadFile("migrations/" + file)
		if readErr != nil {
			t.Fatal(readErr)
		}

		if _, err = pool.Exec(ctx, string(raw), pgx.QueryExecModeSimpleProtocol); err != nil {
			t.Fatal(err)
		}

		if _, err = pool.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, file); err != nil {
			t.Fatal(err)
		}
	}

	_, err = pool.Exec(ctx, `INSERT INTO users(id,name,email,password_hash,role,created_at) VALUES
 ('77777777-7777-4777-8777-777777777701','Pessoa única','unique@example.test','test','member',now()),
 ('77777777-7777-4777-8777-777777777702','QA duplicado','duplicate1@example.test','test','member',now()),
 ('77777777-7777-4777-8777-777777777703','QA duplicado','duplicate2@example.test','test','member',now());
 INSERT INTO work_items(id,title,kind,priority,assignee,reviewer,created_at,updated_at) VALUES
 ('77777777-7777-4777-8777-777777777710','Demanda anterior à migração','improvement','medium',' Pessoa única ','QA duplicado',now(),now()),
 ('77777777-7777-4777-8777-777777777711','Demanda sem conta cadastrada','rewrite','low','Sem cadastro','',now(),now());
 INSERT INTO work_item_transitions(id,demand_id,from_status,to_status,actor,created_at) VALUES
 ('77777777-7777-4777-8777-777777777720','77777777-7777-4777-8777-777777777710','backlog','in_progress','Pessoa única (77777777-7777-4777-8777-777777777701)',now());`, pgx.QueryExecModeSimpleProtocol)
	if err != nil {
		t.Fatal(err)
	}

	if err = db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	if err = db.Migrate(ctx); err != nil {
		t.Fatal("migration not idempotent:", err)
	}

	var assigneeID, reviewerID, name, reviewerName, kind string

	err = pool.QueryRow(ctx, `SELECT COALESCE(assignee_id::text,''),COALESCE(reviewer_id::text,''),assignee,reviewer,kind FROM work_items WHERE id='77777777-7777-4777-8777-777777777710'`).Scan(&assigneeID, &reviewerID, &name, &reviewerName, &kind)
	if err != nil {
		t.Fatal(err)
	}

	if assigneeID != "77777777-7777-4777-8777-777777777701" || reviewerID != "" || name != " Pessoa única " || reviewerName != "QA duplicado" || kind != "improvement" {
		t.Fatal("legacy identity or category incorrectly migrated")
	}

	var actorID string
	if err = pool.QueryRow(ctx, `SELECT actor_id FROM work_item_transitions LIMIT 1`).Scan(&actorID); err != nil || actorID != assigneeID {
		t.Fatal("registered audit identity was lost", err)
	}

	var unknownID string
	if err = pool.QueryRow(ctx, `SELECT COALESCE(assignee_id::text,'') FROM work_items WHERE id='77777777-7777-4777-8777-777777777711'`).Scan(&unknownID); err != nil || unknownID != "" {
		t.Fatal("unregistered user was invented", err)
	}

	assertCatalogSeed(t, db)
}

func assertCatalogSeed(t *testing.T, db *postgres.Client) {
	t.Helper()
	ctx, pool := t.Context(), db.Pool

	var err error

	for range 2 {
		if err = db.Seed(ctx); err != nil {
			t.Fatal(err)
		}

		if err = db.SeedPackages(ctx); err != nil {
			t.Fatal(err)
		}
	}

	var seeded, unlinked int
	if err = pool.QueryRow(ctx, `SELECT count(*),count(*) FILTER(WHERE assignee_id IS NULL OR reviewer_id IS NULL) FROM work_items WHERE id::text LIKE '11111111-%' OR id::text LIKE '33333333-%'`).Scan(&seeded, &unlinked); err != nil || seeded != 9 || unlinked != 0 {
		t.Fatal("demo assignments were duplicated or unlinked", err)
	}

	var notices int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM notifications`).Scan(&notices); err != nil || notices != 5 {
		t.Fatal("notification seed duplicated or missed members", err)
	}

	if _, err = pool.Exec(ctx, `UPDATE notifications SET read_at=now()`); err != nil {
		t.Fatal(err)
	}

	if err = db.SeedPackages(ctx); err != nil {
		t.Fatal(err)
	}

	var unread int
	if err = pool.QueryRow(ctx, `SELECT count(*) FILTER(WHERE read_at IS NULL),count(*) FROM notifications`).Scan(&unread, &notices); err != nil || unread != 0 || notices != 5 {
		t.Fatal("seed reset notification read state", err)
	}
}
