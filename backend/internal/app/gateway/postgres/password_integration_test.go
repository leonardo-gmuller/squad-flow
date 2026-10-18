//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/auth"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres"
	authrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/auth"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/testinfra"
)

func TestPasswordResetIsAtomicAndRejectsStaleLogins(t *testing.T) {
	t.Parallel()

	databaseURL := testinfra.Postgres(t)

	ctx := t.Context()

	db, err := postgres.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(db.Close)

	if err = db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	repo := authrepo.NewAuthRepository(db.Pool)
	id, _ := identifier.NewUUID()

	user := entity.User{ID: id, Name: "Password race test", Email: id + "@example.com", PasswordHash: "previous-hash", CreatedAt: time.Now().UTC()}
	if err = repo.CreateUser(ctx, user); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	tokenHash := id + "-reset"
	rollbackToken := id + "-rollback-reset"

	if err = repo.CreatePasswordReset(ctx, id, id, tokenHash, now, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	results := make(chan error, 2)

	for range 2 {
		go func() { results <- repo.ResetPassword(ctx, id, tokenHash, "new-hash", now) }()
	}

	succeeded := 0

	for range 2 {
		err = <-results
		if err == nil {
			succeeded++
		} else if !errors.Is(err, auth.ErrInvalidReset) {
			t.Fatal(err)
		}
	}

	if succeeded != 1 {
		t.Fatalf("reset consumed %d times", succeeded)
	}

	sessionID, _ := identifier.NewUUID()

	session := entity.Session{ID: sessionID, UserID: id, ExpiresAt: now.Add(time.Hour)}
	if err = repo.CreateSession(ctx, session, id+"-stale-refresh", "previous-hash"); !errors.Is(err, auth.ErrUnauthorized) {
		t.Fatalf("stale password created a session: %v", err)
	}

	if err = repo.CreateSession(ctx, session, id+"-current-refresh", "new-hash"); err != nil {
		t.Fatal(err)
	}

	if err = repo.ChangePassword(ctx, id, id, "final-hash", "new-hash"); err != nil {
		t.Fatal(err)
	}

	if _, err = repo.SessionUser(ctx, sessionID, now); !errors.Is(err, auth.ErrUnauthorized) {
		t.Fatal("session survived credential change")
	}

	if err = repo.ChangePassword(ctx, id, id, "stale-change", "new-hash"); err == nil {
		t.Fatal("stale current password changed credentials")
	}

	stored, err := repo.ByEmail(context.WithoutCancel(ctx), user.Email)
	if err != nil || stored.PasswordHash != "final-hash" {
		t.Fatal("credential state is inconsistent")
	}

	// A failure in the last write (audit actor FK) must roll back the password,
	// session revocation and invalidation of an outstanding reset token.
	freshID, _ := identifier.NewUUID()
	if err = repo.CreateSession(ctx, entity.Session{ID: freshID, UserID: id, ExpiresAt: now.Add(time.Hour)}, id+"-rollback-refresh", "final-hash"); err != nil {
		t.Fatal(err)
	}

	if err = repo.CreatePasswordReset(ctx, id, id, rollbackToken, now, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	missingActor, _ := identifier.NewUUID()
	if err = repo.ChangePassword(ctx, id, missingActor, "must-not-persist", "final-hash"); err == nil {
		t.Fatal("invalid audit actor unexpectedly committed")
	}

	stored, err = repo.ByEmail(ctx, user.Email)
	if err != nil || stored.PasswordHash != "final-hash" {
		t.Fatal("password changed despite rollback")
	}

	if _, err = repo.SessionUser(ctx, freshID, now); err != nil {
		t.Fatal("session was revoked despite rollback")
	}

	if _, err = repo.PasswordResetUser(ctx, rollbackToken, now); err != nil {
		t.Fatal("reset token was consumed despite rollback")
	}
}
