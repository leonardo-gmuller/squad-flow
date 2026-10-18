package authrepo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/auth"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *AuthRepository) Rotate(ctx context.Context, oldHash, newHash string, now time.Time) (entity.Session, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return entity.Session{}, erring.Wrap(err, "authrepo.Rotate")
	}

	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	queries := r.WithTx(tx)

	row, err := queries.LockRefreshSession(ctx, oldHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Session{}, auth.ErrUnauthorized
	}

	if err != nil {
		return entity.Session{}, erring.Wrap(err, "authrepo.Rotate")
	}

	session := toSession(row)
	if row.Consumed || row.Revoked || !session.ExpiresAt.After(now) {
		if err = queries.RevokeSession(ctx, session.ID); err != nil {
			return session, erring.Wrap(err, "authrepo.Rotate")
		}

		if err = tx.Commit(ctx); err != nil {
			return session, erring.Wrap(err, "authrepo.Rotate")
		}

		return session, auth.ErrUnauthorized
	}

	if err = queries.ConsumeRefreshToken(ctx, oldHash); err != nil {
		return session, erring.Wrap(err, "authrepo.Rotate")
	}

	if err = queries.RotateRefreshToken(ctx, sqlc.RotateRefreshTokenParams{TokenHash: newHash, SessionID: session.ID}); err != nil {
		return session, erring.Wrap(err, "authrepo.Rotate")
	}

	return session, erring.Wrap(tx.Commit(ctx), "authrepo.Rotate")
}

func (r *AuthRepository) Revoke(ctx context.Context, hash string) error {
	return erring.Wrap(r.RevokeRefreshSession(ctx, hash), "authrepo.Revoke")
}
