package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

func TestErrorContractPreservesCodesAndHidesInternalDetails(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name          string
		err           error
		status        int
		code, message string
	}{
		{"credentials", fmt.Errorf("auth.Login private context: %w", erring.ErrInvalidCredentials), 401, "auth:invalid-credentials", "Email ou senha inválidos."},
		{"validation", fmt.Errorf("release.Create: %w", erring.ErrInvalidInput.WithMessage("Informe a versão.")), 400, "request:invalid", "Informe a versão."},
		{"forbidden", erring.ErrAdminRequired, 403, "auth:admin-required", erring.ErrAdminRequired.Message},
		{"conflict", erring.ErrEmailAlreadyUsed, 409, "auth:email-already-used", erring.ErrEmailAlreadyUsed.Message},
		{"missing", erring.ErrNotFound, 404, "resource:not-found", erring.ErrNotFound.Message},
		{"limit", erring.ErrTooManyRequests, 429, "request:too-many-requests", erring.ErrTooManyRequests.Message},
		{"unexpected", errors.New("private SQL password details"), 500, "oops:internal-server-error", "Ocorreu um erro inesperado. Tente novamente mais tarde."},
		{"unmapped", erring.NewAppError("private:unknown", "private upstream payload"), 500, "oops:internal-server-error", "Ocorreu um erro inesperado. Tente novamente mais tarde."},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			recorded := httptest.NewRecorder()
			Send(recorded, AppError(test.err))

			var body Error
			if err := json.Unmarshal(recorded.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}

			if recorded.Code != test.status || body.Code != test.code || body.Message != test.message || !strings.HasPrefix(body.Type, "srn:error:") || strings.Contains(recorded.Body.String(), "private") {
				t.Fatalf("unexpected error contract: status=%d body=%s", recorded.Code, recorded.Body.String())
			}
		})
	}
}

func TestHandleWritesHeadersMultipleCookiesAndNoContent(t *testing.T) {
	t.Parallel()

	handler := Handle(func(_ *http.Request) *Response {
		return NoContent().WithHeaders(map[string]string{"X-Test": "ok"}).WithCookies(
			&http.Cookie{Name: "access", Value: "test-access", HttpOnly: true},
			&http.Cookie{Name: "refresh", Value: "test-refresh", HttpOnly: true},
		)
	})
	recorded := httptest.NewRecorder()
	handler.ServeHTTP(recorded, httptest.NewRequest(http.MethodPost, "/", nil))

	if recorded.Code != http.StatusNoContent || recorded.Body.Len() != 0 || recorded.Header().Get("X-Test") != "ok" || len(recorded.Header().Values("Set-Cookie")) != 2 {
		t.Fatal("response metadata or empty-body contract changed")
	}
}

func TestDecodeRejectsMalformedUnknownAndMultipleDocuments(t *testing.T) {
	t.Parallel()

	for _, body := range []string{`null`, `[]`, `{"password":`, `{"password":"secret","unknown":true}`, `{"password":"one"} {"password":"two"}`} {
		var input struct {
			Password string `json:"password"`
		}

		err := Decode(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)), &input)
		if !errors.Is(err, erring.ErrInvalidInput) {
			t.Fatalf("invalid JSON accepted: %v", err)
		}
	}

	var input struct {
		Password string `json:"password"`
	}

	err := Decode(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"password":"`+strings.Repeat("a", 1<<20)+`"}`)), &input)
	if !errors.Is(err, erring.ErrPayloadTooLarge) {
		t.Fatalf("oversized body accepted: %v", err)
	}
}
