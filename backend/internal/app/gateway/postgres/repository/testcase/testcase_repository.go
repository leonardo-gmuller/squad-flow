package testcaserepo

import (
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/uow"
)

type TestCaseRepository struct{ *sqlc.Queries }

var _ demand.TestRepository = (*TestCaseRepository)(nil)

func NewTestCaseRepository(db uow.DBTX) *TestCaseRepository {
	return &TestCaseRepository{Queries: sqlc.New(db)}
}
