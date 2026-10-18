package postgres

import (
	"context"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/security"
)

func (s *Client) Seed(ctx context.Context) error {
	if err := s.SeedMembers(ctx); err != nil {
		return err
	}

	return erring.Wrap(sqlc.New(s.Pool).SeedDemands(ctx), "postgres.Seed")
}

// SeedPackages illustrates fictional OS deliveries split across registered owners.
func (s *Client) SeedPackages(ctx context.Context) error {
	if err := s.SeedMembers(ctx); err != nil {
		return err
	}

	queries := sqlc.New(s.Pool)
	if err := queries.SeedPackages(ctx); err != nil {
		return erring.Wrap(err, "postgres.SeedPackages")
	}

	if err := queries.SeedPackageDemands(ctx); err != nil {
		return erring.Wrap(err, "postgres.SeedPackages")
	}

	return s.SeedNotifications(ctx)
}

// Demo accounts exist only when DEMO_SEED is enabled (forbidden in production).
func (s *Client) SeedMembers(ctx context.Context) error {
	hash, err := (security.Passwords{}).Hash("SquadFlow-Demo-2026!")
	if err != nil {
		return erring.Wrap(err, "postgres.SeedMembers.password")
	}

	queries := sqlc.New(s.Pool)
	if err = queries.SeedMembers(ctx, hash); err != nil {
		return erring.Wrap(err, "postgres.SeedMembers.insert")
	}

	return erring.Wrap(queries.LinkSeedAssignments(ctx), "postgres.SeedMembers")
}

func (s *Client) SeedNotifications(ctx context.Context) error {
	return erring.Wrap(sqlc.New(s.Pool).SeedNotifications(ctx), "postgres.SeedNotifications")
}
