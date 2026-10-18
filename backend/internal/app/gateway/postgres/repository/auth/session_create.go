package authrepo

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/auth"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *AuthRepository) CreateSession(ctx context.Context, session entity.Session, hash, passwordHash string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return erring.Wrap(err, "authrepo.CreateSession")
	}

	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	queries := r.WithTx(tx)
	// Serialize login with credential changes: a stale password cannot recreate a session.
	currentHash, err := queries.LockUserPassword(ctx, session.UserID)
	if err != nil {
		return erring.Wrap(err, "authrepo.CreateSession")
	}

	if currentHash != passwordHash {
		return auth.ErrUnauthorized
	}

	if err = queries.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:        session.ID,
		UserID:    session.UserID,
		ExpiresAt: session.ExpiresAt,
	}); err != nil {
		return erring.Wrap(err, "authrepo.CreateSession")
	}

	if err = queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{TokenHash: hash, SessionID: session.ID}); err != nil {
		return erring.Wrap(err, "authrepo.CreateSession")
	}

	return erring.Wrap(tx.Commit(ctx), "authrepo.CreateSession")
}
