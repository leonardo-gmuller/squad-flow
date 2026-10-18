package authrepo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/auth"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *AuthRepository) CreatePasswordReset(ctx context.Context, userID, actorID, tokenHash string, now, expires time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return erring.Wrap(err, "authrepo.CreatePasswordReset")
	}

	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	queries := r.WithTx(tx)

	if _, err = lockPasswordUser(ctx, queries, userID); err != nil {
		return err
	}

	if err = queries.ConsumeUserResetTokensAt(ctx, sqlc.ConsumeUserResetTokensAtParams{UserID: userID, Now: &now}); err != nil {
		return erring.Wrap(err, "authrepo.CreatePasswordReset")
	}

	if err = queries.CreatePasswordResetToken(ctx, sqlc.CreatePasswordResetTokenParams{
		TokenHash: tokenHash,
		UserID:    userID,
		ActorID:   actorID,
		CreatedAt: now,
		ExpiresAt: expires,
	}); err != nil {
		return erring.Wrap(err, "authrepo.CreatePasswordReset")
	}

	if err = queries.CreateResetLinkEvent(ctx, sqlc.CreateResetLinkEventParams{
		UserID:    userID,
		ActorID:   actorID,
		CreatedAt: now,
	}); err != nil {
		return erring.Wrap(err, "authrepo.CreatePasswordReset")
	}

	return erring.Wrap(tx.Commit(ctx), "authrepo.CreatePasswordReset")
}

func (r *AuthRepository) PasswordResetUser(ctx context.Context, tokenHash string, now time.Time) (string, error) {
	userID, err := r.GetPasswordResetUser(ctx, sqlc.GetPasswordResetUserParams{TokenHash: tokenHash, Now: now})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", auth.ErrInvalidReset
	}

	return userID, erring.Wrap(err, "authrepo.PasswordResetUser")
}

func (r *AuthRepository) ResetPassword(ctx context.Context, userID, tokenHash, passwordHash string, now time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return erring.Wrap(err, "authrepo.ResetPassword")
	}

	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	queries := r.WithTx(tx)

	if _, err = lockPasswordUser(ctx, queries, userID); err != nil {
		return err
	}

	affected, err := queries.ConsumePasswordResetToken(ctx, sqlc.ConsumePasswordResetTokenParams{
		TokenHash: tokenHash,
		UserID:    userID,
		Now:       &now,
	})
	if err != nil {
		return erring.Wrap(err, "authrepo.ResetPassword")
	}

	if affected != 1 {
		return auth.ErrInvalidReset
	}

	if err = writePassword(ctx, queries, userID, userID, passwordHash, "password_reset"); err != nil {
		return err
	}

	return erring.Wrap(tx.Commit(ctx), "authrepo.ResetPassword")
}
