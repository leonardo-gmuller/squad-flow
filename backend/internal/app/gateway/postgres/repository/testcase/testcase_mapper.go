package testcaserepo

import (
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func createTestCaseParams(item entity.TestCase) sqlc.CreateTestCaseParams {
	return sqlc.CreateTestCaseParams{
		ID:        item.ID,
		DemandID:  item.DemandID,
		Cycle:     item.Cycle,
		Title:     item.Title,
		Steps:     item.Steps,
		Expected:  item.Expected,
		Actual:    item.Actual,
		Result:    item.Result,
		Author:    item.Author,
		CreatedAt: item.CreatedAt,
		AuthorID:  item.AuthorID,
	}
}
