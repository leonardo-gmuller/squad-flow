package packagerepo

import (
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func toPackage(row sqlc.GetPackageRow) entity.WorkPackage {
	return entity.WorkPackage{
		ID:          row.ID,
		Reference:   row.Reference,
		Title:       row.Title,
		Description: row.Description,
		Initiative:  row.Initiative,
		Kind:        entity.Kind(row.Kind),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		Total:       row.Total,
		Done:        row.Done,
		InQA:        row.InQA,
	}
}

func createPackageParams(item entity.WorkPackage) sqlc.CreatePackageParams {
	return sqlc.CreatePackageParams{
		ID:          item.ID,
		Reference:   item.Reference,
		Title:       item.Title,
		Description: item.Description,
		Initiative:  item.Initiative,
		Kind:        string(item.Kind),
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}
