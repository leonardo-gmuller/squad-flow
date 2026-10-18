//go:build integration

// Package testinfra owns disposable infrastructure for integration tests.
package testinfra

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
)

// Postgres returns a new database owned by this test, using an ephemeral port.
func Postgres(t *testing.T) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()

	container, err := pgcontainer.Run(ctx, "postgres:17-alpine",
		pgcontainer.WithDatabase("squadflow_test"),
		pgcontainer.WithUsername("squadflow"),
		pgcontainer.WithPassword("testcontainers-only-password"),
		pgcontainer.BasicWaitStrategies(),
	)
	testcontainers.CleanupContainer(t, container)

	if err != nil {
		t.Fatalf("start PostgreSQL test container (Docker must be running): %v", err)
	}

	connection, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get PostgreSQL test connection: %v", err)
	}

	return connection
}

func Redis(t *testing.T) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()

	container, err := rediscontainer.Run(ctx, "redis:7.4-alpine")
	testcontainers.CleanupContainer(t, container)

	if err != nil {
		t.Fatalf("start Redis test container (Docker must be running): %v", err)
	}

	connection, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("get Redis test connection: %v", err)
	}

	return connection
}
