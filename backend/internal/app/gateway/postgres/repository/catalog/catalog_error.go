package catalogrepo

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

func writeError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return erring.ErrNotFound
	}

	var pe *pgconn.PgError
	if errors.As(err, &pe) && pe.Code == "23505" {
		return erring.ErrConflict.WithMessage("já existe uma esteira com esse nome")
	}

	return erring.Wrap(err, "catalogrepo.write")
}
