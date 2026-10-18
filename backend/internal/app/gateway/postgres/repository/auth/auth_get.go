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

func (r *AuthRepository) ByEmail(ctx context.Context, email string) (entity.User, error) {
	row, err := r.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.User{}, auth.ErrUnauthorized
	}

	if err != nil {
		return entity.User{}, erring.Wrap(err, "authrepo.ByEmail")
	}

	return toUser(row), nil
}

func (r *AuthRepository) Users(ctx context.Context) ([]entity.User, error) {
	rows, err := r.ListUsers(ctx)
	if err != nil {
		return nil, erring.Wrap(err, "authrepo.Users")
	}

	items := make([]entity.User, 0, len(rows))
	for _, row := range rows {
		items = append(items, toUser(sqlc.GetUserByEmailRow(row)))
	}

	return items, nil
}

func (r *AuthRepository) SessionUser(ctx context.Context, id string, now time.Time) (entity.User, error) {
	row, err := r.GetSessionUser(ctx, sqlc.GetSessionUserParams{ID: id, Now: now})
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.User{}, auth.ErrUnauthorized
	}

	if err != nil {
		return entity.User{}, erring.Wrap(err, "authrepo.SessionUser")
	}

	return toUser(sqlc.GetUserByEmailRow(row)), nil
}
