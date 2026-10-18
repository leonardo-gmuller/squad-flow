// Package uow implements a transaction-scoped repository registry.
// Unlike an instance with a mutable tx field, the registry can serve
// concurrent HTTP requests safely; each Do callback gets an isolated scope.
package uow

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
}

type RepoFactory func(DBTX) any

type PGX struct {
	pool         *pgxpool.Pool
	mu           sync.RWMutex
	repositories map[string]RepoFactory
}

type TxScope struct {
	tx        pgx.Tx
	factories map[string]RepoFactory
}

func NewPGX(pool *pgxpool.Pool) *PGX {
	return &PGX{pool: pool, repositories: make(map[string]RepoFactory)}
}

func (u *PGX) Register(name string, factory RepoFactory) error {
	if name == "" || factory == nil {
		return errors.New("uow: invalid repository registration")
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	if _, exists := u.repositories[name]; exists {
		return fmt.Errorf("uow: repository %q already registered", name)
	}

	u.repositories[name] = factory

	return nil
}

func (u *PGX) GetRepositoryFromPool(name string) (any, error) {
	u.mu.RLock()
	factory, ok := u.repositories[name]
	u.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("uow: repository %q not registered", name)
	}

	return factory(u.pool), nil
}

func (s *TxScope) GetRepositoryTx(name string) (any, error) {
	factory, ok := s.factories[name]
	if !ok {
		return nil, fmt.Errorf("uow: repository %q not registered", name)
	}

	return factory(s.tx), nil
}

func RepoTx[T any](scope *TxScope, name string) (T, error) {
	var zero T

	repo, err := scope.GetRepositoryTx(name)
	if err != nil {
		return zero, err
	}

	typed, ok := repo.(T)
	if !ok {
		return zero, fmt.Errorf("uow: repository %q has different type", name)
	}

	return typed, nil
}

func RepoPool[T any](u *PGX, name string) (T, error) {
	var zero T

	repo, err := u.GetRepositoryFromPool(name)
	if err != nil {
		return zero, err
	}

	typed, ok := repo.(T)
	if !ok {
		return zero, fmt.Errorf("uow: repository %q has different type", name)
	}

	return typed, nil
}

// Do creates one transaction per call. Rollback is deferred even if fn panics.
// Factory registry is snapshotted while locked, preventing registration races.
func (u *PGX) Do(ctx context.Context, fn func(context.Context, *TxScope) error) error {
	if u == nil || u.pool == nil {
		return errors.New("uow: connection pool unavailable")
	}

	u.mu.RLock()

	factories := make(map[string]RepoFactory, len(u.repositories))
	for k, v := range u.repositories {
		factories[k] = v
	}
	u.mu.RUnlock()

	tx, err := u.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("uow.begin: %w", err)
	}

	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	if err := fn(ctx, &TxScope{tx: tx, factories: factories}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("uow.commit: %w", err)
	}

	return nil
}
