package catalogrepo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

// Locks serialize assignment with deactivation inside the same write transaction.
func ValidateKind(ctx context.Context, queries *sqlc.Queries, kind entity.Kind, existing bool) error {
	active, err := queries.GetWorkstreamActiveForShare(ctx, string(kind))
	if errors.Is(err, pgx.ErrNoRows) {
		return erring.ErrInvalidInput.WithMessage("esteira não cadastrada")
	}

	if err != nil {
		return erring.Wrap(err, "catalogrepo.ValidateKind")
	}

	if !active && !existing {
		return erring.ErrConflict.WithMessage("esta esteira está desativada")
	}

	return nil
}

func ResolveMember(ctx context.Context, queries *sqlc.Queries, id string) (string, error) {
	if id == "" {
		return "", nil
	}

	if !identifier.IsUUID(id) {
		return "", erring.ErrInvalidInput.WithMessage("selecione um usuário cadastrado")
	}

	name, err := queries.GetMemberNameForKeyShare(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", erring.ErrInvalidInput.WithMessage("usuário não cadastrado")
	}

	return name, erring.Wrap(err, "catalogrepo.ResolveMember")
}
