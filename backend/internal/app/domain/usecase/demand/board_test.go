package demand

import (
	"errors"
	"testing"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

func TestBoardRejectsInvalidFilters(t *testing.T) {
	t.Parallel()

	for _, filter := range []BoardFilter{{ReleaseID: "invalid"}, {Owner: "123"}, {Archive: "deleted"}, {Kind: "invalid kind!"}, {Status: "closed"}, {Limit: 101}, {Limit: -1}, {Offset: -1}, {Offset: 1}} {
		_, err := NewBoardUseCase(nil).Board(t.Context(), filter)
		if !errors.Is(err, erring.ErrInvalidInput) {
			t.Fatalf("invalid filter accepted: %+v %v", filter, err)
		}
	}
}
