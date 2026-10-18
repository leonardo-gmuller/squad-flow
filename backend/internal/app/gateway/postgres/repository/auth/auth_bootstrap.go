package authrepo

import (
	"context"
	"errors"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *AuthRepository) Bootstrap(ctx context.Context, user entity.User) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return erring.Wrap(err, "authrepo.Bootstrap")
	}

	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	queries := r.WithTx(tx)
	if err = queries.LockAdminBootstrap(ctx); err != nil {
		return erring.Wrap(err, "authrepo.Bootstrap")
	}

	exists, err := queries.AdminExists(ctx)
	if err != nil {
		return erring.Wrap(err, "authrepo.Bootstrap")
	}

	if !exists {
		if user.ID == "" {
			return errors.New("configure ADMIN_EMAIL e ADMIN_PASSWORD para criar o único administrador")
		}

		if err = queries.CreateAdmin(ctx, sqlc.CreateAdminParams{
			ID:           user.ID,
			Name:         user.Name,
			Email:        user.Email,
			PasswordHash: user.PasswordHash,
			CreatedAt:    user.CreatedAt,
		}); err != nil {
			return erring.Wrap(err, "authrepo.Bootstrap")
		}
	}

	return erring.Wrap(tx.Commit(ctx), "authrepo.Bootstrap")
}
