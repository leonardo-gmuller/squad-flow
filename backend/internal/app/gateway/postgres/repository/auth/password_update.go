package authrepo

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

func (r *AuthRepository) ChangePassword(ctx context.Context, userID, actorID, passwordHash, expectedHash string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return erring.Wrap(err, "authrepo.ChangePassword")
	}

	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	queries := r.WithTx(tx)

	currentHash, err := lockPasswordUser(ctx, queries, userID)
	if err != nil {
		return err
	}

	if expectedHash != "" && expectedHash != currentHash {
		return erring.ErrConflict.WithMessage("a senha já foi alterada; entre novamente")
	}

	if err = writePassword(ctx, queries, userID, actorID, passwordHash, "password_changed"); err != nil {
		return err
	}

	return erring.Wrap(tx.Commit(ctx), "authrepo.ChangePassword")
}
