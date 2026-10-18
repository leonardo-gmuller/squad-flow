package demand

import (
	"context"
	"strings"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

type BoardFilter struct {
	ReleaseID string
	Archive   string
	Kind      string
	Owner     string
	Search    string
	Status    string
	Limit     int
	Offset    int
}

type BoardColumn struct {
	Status entity.Status   `json:"status"`
	Total  int             `json:"total"`
	Items  []entity.Demand `json:"items"`
}

type Board struct {
	Columns []BoardColumn `json:"columns"`
	Owners  []string      `json:"owners"`
}

//nolint:iface // Repository reads and use case validation are distinct architectural boundaries.
type BoardRepository interface {
	Board(ctx context.Context, filter BoardFilter) (Board, error)
}
type BoardUseCase struct{ repo BoardRepository }

func NewBoardUseCase(repo BoardRepository) *BoardUseCase { return &BoardUseCase{repo: repo} }

func (s *BoardUseCase) Board(ctx context.Context, filter BoardFilter) (Board, error) {
	if filter.Archive == "" {
		filter.Archive = "active"
	}

	filter.Search = strings.TrimSpace(filter.Search)
	filter.ReleaseID = strings.ToLower(filter.ReleaseID)

	if filter.Limit == 0 {
		filter.Limit = 20
	}

	if (filter.ReleaseID != "" && filter.ReleaseID != "none" && !identifier.IsUUID(filter.ReleaseID)) ||
		(filter.Archive != "active" && filter.Archive != "archived" && filter.Archive != "all") ||
		(filter.Owner != "" && !identifier.IsUUID(filter.Owner)) ||
		(filter.Kind != "" && !entity.IsKind(entity.Kind(filter.Kind))) ||
		(filter.Status != "" && !entity.IsStatus(entity.Status(filter.Status))) ||
		filter.Limit < 1 || filter.Limit > 100 || filter.Offset < 0 || len(filter.Search) > 300 || len(filter.Owner) > 120 ||
		(filter.Offset > 0 && filter.Status == "") {
		return Board{}, erring.ErrInvalidInput.WithMessage("filtros do quadro inválidos")
	}

	result, err := s.repo.Board(ctx, filter)

	return result, erring.Wrap(err, "demand.Board")
}
