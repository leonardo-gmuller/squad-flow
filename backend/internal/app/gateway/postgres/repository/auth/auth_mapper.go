package authrepo

import (
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func toUser(row sqlc.GetUserByEmailRow) entity.User {
	return entity.User{
		ID:           row.ID,
		Name:         row.Name,
		Email:        row.Email,
		Role:         row.Role,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
	}
}

func toSession(row sqlc.LockRefreshSessionRow) entity.Session {
	return entity.Session{
		ID:        row.ID,
		UserID:    row.UserID,
		ExpiresAt: row.ExpiresAt,
	}
}
