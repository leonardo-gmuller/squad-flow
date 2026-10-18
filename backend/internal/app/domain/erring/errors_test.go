package erring

import (
	"errors"
	"strings"
	"testing"
)

func TestWrapPreservesNil(t *testing.T) {
	t.Parallel()

	if Wrap(nil, "save") != nil {
		t.Fatal("a successful operation must keep a nil error")
	}
}

func TestWrapPreservesCauseAndOperation(t *testing.T) {
	t.Parallel()

	cause := errors.New("storage unavailable")
	wrapped := Wrap(cause, "repository.save")

	if !errors.Is(wrapped, cause) || !strings.Contains(wrapped.Error(), "repository.save") {
		t.Fatalf("operation or cause lost: %v", wrapped)
	}
}
