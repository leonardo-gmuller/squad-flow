package auth

import (
	"context"
	"errors"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

var ErrInvalidReset = erring.ErrInvalidResetToken

type PasswordReset struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (s *Service) AdminChangePassword(ctx context.Context, actor entity.User, userID, password string) error {
	if actor.Role != adminRole {
		return ErrForbidden
	}

	if !identifier.IsUUID(userID) {
		return erring.ErrInvalidInput.WithMessage("usuário inválido")
	}

	if err := validatePassword(password, "senha"); err != nil {
		return err
	}

	hash, err := s.passwords.Hash(password)
	if err != nil {
		return erring.Wrap(err, "auth.AdminChangePassword")
	}

	return erring.Wrap(s.passwordRepo.ChangePassword(ctx, userID, actor.ID, hash, ""), "auth.AdminChangePassword")
}

func (s *Service) ChangePassword(ctx context.Context, actor entity.User, current, password string) error {
	if len(current) > 72 {
		return erring.ErrInvalidCurrentPassword
	}

	user, err := s.repo.ByEmail(ctx, actor.Email)
	if err != nil {
		return erring.Wrap(err, "auth.ChangePassword")
	}

	if user.ID != actor.ID || !s.passwords.Match(user.PasswordHash, current) {
		return erring.ErrInvalidCurrentPassword
	}

	if err = validatePassword(password, "nova senha"); err != nil {
		return err
	}

	hash, err := s.passwords.Hash(password)
	if err != nil {
		return erring.Wrap(err, "auth.ChangePassword")
	}

	return erring.Wrap(s.passwordRepo.ChangePassword(ctx, actor.ID, actor.ID, hash, user.PasswordHash), "auth.ChangePassword")
}

func (s *Service) CreatePasswordReset(ctx context.Context, actor entity.User, userID string) (PasswordReset, error) {
	if actor.Role != adminRole {
		return PasswordReset{}, ErrForbidden
	}

	if !identifier.IsUUID(userID) {
		return PasswordReset{}, erring.ErrInvalidInput.WithMessage("usuário inválido")
	}

	raw, _, err := s.tokens.Refresh()
	if err != nil {
		return PasswordReset{}, erring.Wrap(err, "auth.CreatePasswordReset")
	}

	now := time.Now().UTC()
	result := PasswordReset{Token: raw, ExpiresAt: now.Add(30 * time.Minute)}

	err = s.passwordRepo.CreatePasswordReset(ctx, userID, actor.ID, s.resetHash(raw), now, result.ExpiresAt)
	if err != nil {
		return PasswordReset{}, erring.Wrap(err, "auth.CreatePasswordReset")
	}

	return result, nil
}

func (s *Service) ResetPassword(ctx context.Context, raw, password string) (string, error) {
	if len(raw) != 43 {
		return "", ErrInvalidReset
	}

	if err := validatePassword(password, "nova senha"); err != nil {
		return "", err
	}

	// Reject unknown/expired tokens before the expensive password hash. The
	// repository checks again under the user's lock when consuming the token.
	tokenHash := s.resetHash(raw)

	userID, err := s.passwordRepo.PasswordResetUser(ctx, tokenHash, time.Now().UTC())
	if err != nil {
		return "", erring.Wrap(err, "auth.ResetPassword")
	}

	hash, err := s.passwords.Hash(password)
	if err != nil {
		return "", erring.Wrap(err, "auth.ResetPassword")
	}

	err = s.passwordRepo.ResetPassword(ctx, userID, tokenHash, hash, time.Now().UTC())
	if err != nil {
		if errors.Is(err, ErrInvalidReset) {
			return "", ErrInvalidReset
		}

		return "", erring.Wrap(err, "auth.ResetPassword")
	}

	return userID, nil
}

func (s *Service) resetHash(raw string) string {
	return s.tokens.HashRefresh("password-reset:" + raw)
}
