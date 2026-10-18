package auth

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
)

// UseCaseInterface defines the operations exposed by the auth use case.
type UseCaseInterface interface {
	PasswordUseCaseInterface
	Bootstrap(ctx context.Context, name, email, password string) error
	CreateUser(ctx context.Context, actor entity.User, name, email, password string) (entity.User, error)
	Users(ctx context.Context, actor entity.User) ([]entity.User, error)
	Login(ctx context.Context, email, password string) (Result, error)
	Refresh(ctx context.Context, raw string) (Result, error)
	Authenticate(ctx context.Context, raw string) (Principal, error)
	Logout(ctx context.Context, rawRefresh string) error
}

// PasswordUseCaseInterface exposes password management and internal recovery.
type PasswordUseCaseInterface interface {
	AdminChangePassword(ctx context.Context, actor entity.User, userID, password string) error
	ChangePassword(ctx context.Context, actor entity.User, current, password string) error
	CreatePasswordReset(ctx context.Context, actor entity.User, userID string) (PasswordReset, error)
	ResetPassword(ctx context.Context, raw, password string) (string, error)
}

var (
	_ UseCaseInterface         = (*Service)(nil)
	_ PasswordUseCaseInterface = (*Service)(nil)
)
