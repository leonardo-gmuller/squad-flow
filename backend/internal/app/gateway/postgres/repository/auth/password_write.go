package authrepo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func lockPasswordUser(ctx context.Context, queries *sqlc.Queries, userID string) (string, error) {
	hash, err := queries.GetPasswordForUpdate(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", erring.ErrNotFound
	}

	return hash, erring.Wrap(err, "authrepo.lockPasswordUser")
}

// Every credential change is committed together with session/token revocation
// and an audit event. All callers hold the same user lock as CreateSession.
func writePassword(ctx context.Context, queries *sqlc.Queries, userID, actorID, hash, action string) error {
	if err := queries.UpdatePassword(ctx, sqlc.UpdatePasswordParams{ID: userID, PasswordHash: hash}); err != nil {
		return erring.Wrap(err, "authrepo.writePassword")
	}

	if err := queries.RevokeUserSessions(ctx, userID); err != nil {
		return erring.Wrap(err, "authrepo.writePassword")
	}

	if err := queries.ConsumeUserResetTokens(ctx, userID); err != nil {
		return erring.Wrap(err, "authrepo.writePassword")
	}

	err := queries.CreatePasswordEvent(ctx, sqlc.CreatePasswordEventParams{
		UserID:  userID,
		ActorID: actorID,
		Action:  action,
	})

	return erring.Wrap(err, "authrepo.writePassword")
}
