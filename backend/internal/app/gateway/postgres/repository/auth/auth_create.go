package authrepo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func (r *AuthRepository) CreateUser(ctx context.Context, user entity.User) error {
	err := r.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	})

	var pe *pgconn.PgError

	if errors.As(err, &pe) && pe.Code == "23505" {
		return erring.ErrEmailAlreadyUsed
	}

	return erring.Wrap(err, "authrepo.CreateUser")
}
