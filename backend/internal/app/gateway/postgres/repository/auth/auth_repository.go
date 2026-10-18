package authrepo

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/auth"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

type AuthRepository struct {
	*sqlc.Queries
	pool *pgxpool.Pool
}

var (
	_ auth.Repository         = (*AuthRepository)(nil)
	_ auth.PasswordRepository = (*AuthRepository)(nil)
)

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{Queries: sqlc.New(pool), pool: pool}
}
