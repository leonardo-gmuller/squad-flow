package catalogrepo

import (
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func toWorkstream(row sqlc.Workstream) entity.Workstream {
	return entity.Workstream{
		ID:        row.ID,
		Code:      row.Code,
		Name:      row.Name,
		Color:     row.Color,
		Active:    row.Active,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func toMember(row sqlc.ListMembersRow) entity.Member {
	return entity.Member{
		ID:    row.ID,
		Name:  row.Name,
		Email: row.Email,
	}
}
