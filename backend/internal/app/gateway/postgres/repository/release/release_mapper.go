package releaserepo

import (
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func toRelease(row sqlc.ListReleasesRow) entity.Release {
	return entity.Release{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		ReleaseDate: row.ReleaseDate,
		ArchivedAt:  row.ArchivedAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		Total:       row.Total,
		Done:        row.Done,
		InQA:        row.InQA,
	}
}

func createReleaseParams(item entity.Release) sqlc.CreateReleaseParams {
	return sqlc.CreateReleaseParams{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		ReleaseDate: item.ReleaseDate,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}
