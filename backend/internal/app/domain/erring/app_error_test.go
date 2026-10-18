package erring

import (
	"errors"
	"fmt"
	"testing"
)

func TestPublicMessagePreservesCategoryWithoutMutatingSentinel(t *testing.T) {
	t.Parallel()

	message := "Informe uma versão válida."
	err := Wrap(ErrInvalidInput.WithMessage(message), "release.Create")

	if !errors.Is(err, ErrInvalidInput) || ErrInvalidInput.Message == message || Wrap(nil, "noop") != nil {
		t.Fatal("error identity, shared sentinel or nil wrapping changed")
	}

	var public *AppError
	if !errors.As(err, &public) || public.Code != "request:invalid" || public.Message != message {
		t.Fatalf("public error was not preserved: %v", err)
	}

	credentials := fmt.Errorf("auth.Login: %w", ErrInvalidCredentials)
	if !errors.Is(credentials, ErrUnauthorized) || !errors.Is(credentials, ErrInvalidCredentials) {
		t.Fatal("authentication categories lost")
	}
}
