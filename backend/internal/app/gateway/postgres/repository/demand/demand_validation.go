package demandrepo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	catalogrepo "github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/repository/catalog"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

// Holding a share lock until the demand transaction commits prevents assigning an archived release concurrently.
func (r *DemandRepository) validateRelease(ctx context.Context, releaseID, demandID string) (string, error) {
	if releaseID == "" {
		return "", nil
	}

	if !identifier.IsUUID(releaseID) {
		return "", erring.ErrInvalidInput.WithMessage("versão inválida")
	}

	row, err := r.GetReleaseAssignmentForShare(ctx, sqlc.GetReleaseAssignmentForShareParams{ReleaseID: releaseID, DemandID: demandID})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", erring.ErrInvalidInput.WithMessage("versão não encontrada")
	}

	if err != nil {
		return "", erring.Wrap(err, "demandrepo.validateRelease")
	}

	if row.Allowed == nil || !*row.Allowed {
		return "", erring.ErrConflict.WithMessage("a versão está arquivada; restaure-a antes de vincular novas demandas")
	}

	return row.Name, nil
}

func (r *DemandRepository) validateAssignments(ctx context.Context, item *entity.Demand, existing bool) error {
	if err := catalogrepo.ValidateKind(ctx, r.Queries, item.Kind, existing); err != nil {
		return erring.Wrap(err, "demandrepo.validateAssignments.kind")
	}

	if item.Assignee != "" || item.Reviewer != "" {
		return erring.ErrInvalidInput.WithMessage("selecione os responsáveis pelos IDs dos usuários cadastrados")
	}

	var err error

	item.Assignee, err = catalogrepo.ResolveMember(ctx, r.Queries, item.AssigneeID)
	if err != nil {
		return erring.Wrap(err, "demandrepo.validateAssignments.assignee")
	}

	item.Reviewer, err = catalogrepo.ResolveMember(ctx, r.Queries, item.ReviewerID)

	return erring.Wrap(err, "demandrepo.validateAssignments.reviewer")
}
