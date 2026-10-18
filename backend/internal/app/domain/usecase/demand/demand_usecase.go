package demand

import (
	"time"
)

type UseCase struct {
	uow     UnitOfWork
	queries QueryRepository
	now     func() time.Time
}

func NewUseCase(uow UnitOfWork, queries QueryRepository) *UseCase {
	return &UseCase{uow: uow, queries: queries, now: func() time.Time { return time.Now().UTC() }}
}
