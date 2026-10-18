package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

const adminRole = "admin"

var (
	ErrUnauthorized = erring.ErrUnauthorized
	ErrForbidden    = erring.ErrAdminRequired
)

type Repository interface {
	Bootstrap(ctx context.Context, user entity.User) error
	ByEmail(ctx context.Context, email string) (entity.User, error)
	CreateUser(ctx context.Context, user entity.User) error
	Users(ctx context.Context) ([]entity.User, error)
	CreateSession(ctx context.Context, session entity.Session, hash, passwordHash string) error
	Rotate(ctx context.Context, oldHash, newHash string, now time.Time) (entity.Session, error)
	SessionUser(ctx context.Context, id string, now time.Time) (entity.User, error)
	Revoke(ctx context.Context, hash string) error
}

type PasswordRepository interface {
	ChangePassword(ctx context.Context, userID, actorID, passwordHash, expectedHash string) error
	CreatePasswordReset(ctx context.Context, userID, actorID, tokenHash string, now, expires time.Time) error
	PasswordResetUser(ctx context.Context, tokenHash string, now time.Time) (string, error)
	ResetPassword(ctx context.Context, userID, tokenHash, passwordHash string, now time.Time) error
}
type Passwords interface {
	Hash(raw string) (string, error)
	Match(hash, raw string) bool
}
type Tokens interface {
	NewID() (string, error)
	Refresh() (string, string, error)
	HashRefresh(raw string) string
	Sign(session entity.Session, expires time.Time) (string, error)
	Verify(raw string) (string, error)
}
type Result struct {
	User           entity.User
	Access         string
	Refresh        string
	AccessExpires  time.Time
	RefreshExpires time.Time
}
type Principal struct {
	User      entity.User
	SessionID string
}
type Service struct {
	repo                  Repository
	passwordRepo          PasswordRepository
	passwords             Passwords
	tokens                Tokens
	accessTTL, refreshTTL time.Duration
	dummyHash             string
}

func New(repo Repository, passwordRepo PasswordRepository, passwords Passwords, tokens Tokens, accessTTL, refreshTTL time.Duration) *Service {
	dummy, _ := passwords.Hash("invalid-login-timing-placeholder")

	return &Service{repo: repo, passwordRepo: passwordRepo, passwords: passwords, tokens: tokens, accessTTL: accessTTL, refreshTTL: refreshTTL, dummyHash: dummy}
}

func validate(name, email, password string) error {
	if err := validateIdentity(name, email); err != nil {
		return err
	}

	return validatePassword(password, "senha")
}

func validateIdentity(name, email string) error {
	if len(strings.TrimSpace(name)) < 2 || len(name) > 120 {
		return erring.ErrInvalidInput.WithMessage("nome deve conter de 2 a 120 bytes")
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 254 {
		return erring.ErrInvalidInput.WithMessage("email deve ser válido e conter até 254 bytes")
	}

	return nil
}

func validatePassword(password, field string) error {
	if utf8.RuneCountInString(password) < 12 {
		return erring.ErrInvalidInput.WithMessage(field + " deve conter pelo menos 12 caracteres")
	}

	if len(password) > 72 {
		return erring.ErrInvalidInput.WithMessage(field + " muito longa; reduza a quantidade de caracteres")
	}

	return nil
}

func (s *Service) Bootstrap(ctx context.Context, name, email, password string) error {
	if email == "" && password == "" {
		return erring.Wrap(s.repo.Bootstrap(ctx, entity.User{}), "auth.Bootstrap")
	}

	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)

	if err := validateIdentity(name, email); err != nil {
		return fmt.Errorf("configuração de ADMIN_NAME/ADMIN_EMAIL: %w", err)
	}

	if err := validatePassword(password, "ADMIN_PASSWORD"); err != nil {
		return err
	}

	hash, err := s.passwords.Hash(password)
	if err != nil {
		return erring.Wrap(err, "auth.Bootstrap")
	}

	id, err := s.tokens.NewID()
	if err != nil {
		return erring.Wrap(err, "auth.Bootstrap")
	}

	return erring.Wrap(s.repo.Bootstrap(ctx, entity.User{ID: id, Name: name, Email: email, Role: adminRole, PasswordHash: hash, CreatedAt: time.Now().UTC()}), "auth.Bootstrap")
}

func (s *Service) CreateUser(ctx context.Context, actor entity.User, name, email, password string) (entity.User, error) {
	if actor.Role != adminRole {
		return entity.User{}, ErrForbidden
	}

	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)

	if err := validate(name, email, password); err != nil {
		return entity.User{}, err
	}

	hash, err := s.passwords.Hash(password)
	if err != nil {
		return entity.User{}, erring.Wrap(err, "auth.CreateUser")
	}

	id, err := s.tokens.NewID()
	if err != nil {
		return entity.User{}, erring.Wrap(err, "auth.CreateUser")
	}

	u := entity.User{ID: id, Name: name, Email: email, Role: "member", PasswordHash: hash, CreatedAt: time.Now().UTC()}

	return u, erring.Wrap(s.repo.CreateUser(ctx, u), "auth.CreateUser")
}

func (s *Service) Users(ctx context.Context, actor entity.User) ([]entity.User, error) {
	if actor.Role != adminRole {
		return nil, ErrForbidden
	}

	result, queryErr := s.repo.Users(ctx)

	return result, erring.Wrap(queryErr, "auth.Users")
}

func (s *Service) Login(ctx context.Context, email, password string) (Result, error) {
	if len(password) > 72 || len(email) > 254 {
		return Result{}, erring.ErrInvalidCredentials
	}

	user, err := s.repo.ByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, ErrUnauthorized) {
		s.passwords.Match(s.dummyHash, password)

		return Result{}, erring.ErrInvalidCredentials
	}

	if err != nil {
		return Result{}, erring.Wrap(err, "auth.Login")
	}

	if !s.passwords.Match(user.PasswordHash, password) {
		return Result{}, erring.ErrInvalidCredentials
	}

	id, err := s.tokens.NewID()
	if err != nil {
		return Result{}, erring.Wrap(err, "auth.Login")
	}

	now := time.Now().UTC()
	session := entity.Session{ID: id, UserID: user.ID, ExpiresAt: now.Add(s.refreshTTL)}

	raw, hash, err := s.tokens.Refresh()
	if err != nil {
		return Result{}, erring.Wrap(err, "auth.Login")
	}

	access, err := s.tokens.Sign(session, now.Add(s.accessTTL))
	if err != nil {
		return Result{}, erring.Wrap(err, "auth.Login")
	}

	if err = s.repo.CreateSession(ctx, session, hash, user.PasswordHash); err != nil {
		return Result{}, erring.Wrap(err, "auth.Login")
	}

	return Result{user, access, raw, now.Add(s.accessTTL), session.ExpiresAt}, nil
}

func (s *Service) Refresh(ctx context.Context, raw string) (Result, error) {
	if raw == "" || len(raw) > 128 {
		return Result{}, ErrUnauthorized
	}

	next, hash, err := s.tokens.Refresh()
	if err != nil {
		return Result{}, erring.Wrap(err, "auth.Refresh")
	}

	now := time.Now().UTC()

	session, err := s.repo.Rotate(ctx, s.tokens.HashRefresh(raw), hash, now)
	if err != nil {
		return Result{}, erring.Wrap(err, "auth.Refresh")
	}

	user, err := s.repo.SessionUser(ctx, session.ID, now)
	if err != nil {
		return Result{}, erring.Wrap(err, "auth.Refresh")
	}

	access, err := s.tokens.Sign(session, now.Add(s.accessTTL))
	if err != nil {
		return Result{}, erring.Wrap(err, "auth.Refresh")
	}

	return Result{user, access, next, now.Add(s.accessTTL), session.ExpiresAt}, nil
}

func (s *Service) Authenticate(ctx context.Context, raw string) (Principal, error) {
	id, err := s.tokens.Verify(raw)
	if err != nil {
		return Principal{}, ErrUnauthorized
	}

	u, err := s.repo.SessionUser(ctx, id, time.Now().UTC())

	return Principal{u, id}, erring.Wrap(err, "auth.Authenticate")
}

func (s *Service) Logout(ctx context.Context, rawRefresh string) error {
	if rawRefresh == "" {
		return nil
	}

	return erring.Wrap(s.repo.Revoke(ctx, s.tokens.HashRefresh(rawRefresh)), "auth.Logout")
}
