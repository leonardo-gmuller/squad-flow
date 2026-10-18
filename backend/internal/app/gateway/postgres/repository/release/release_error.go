package releaserepo

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

func writeError(err error) error {
	var pe *pgconn.PgError
	if errors.As(err, &pe) && pe.Code == "23505" {
		return erring.ErrConflict.WithMessage("já existe uma versão com esse nome")
	}

	return erring.Wrap(err, "releaserepo.write")
}
