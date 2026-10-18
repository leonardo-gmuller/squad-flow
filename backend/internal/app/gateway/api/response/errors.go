package response

import (
	"errors"
	"net/http"
	"strings"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

type Error struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func StatusCodeFromError(err error) int {
	switch {
	case errors.Is(err, erring.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, erring.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, erring.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, erring.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, erring.ErrConflict), errors.Is(err, erring.ErrInvalidTransition):
		return http.StatusConflict
	case errors.Is(err, erring.ErrTooManyRequests):
		return http.StatusTooManyRequests
	case errors.Is(err, erring.ErrPayloadTooLarge):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, erring.ErrMethodNotAllowed):
		return http.StatusMethodNotAllowed
	case errors.Is(err, erring.ErrServiceUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func AppError(err error) *Response {
	status := StatusCodeFromError(err)
	public := Error{Type: errorType(status), Code: "oops:internal-server-error", Message: "Ocorreu um erro inesperado. Tente novamente mais tarde."}

	var appError *erring.AppError
	if status != http.StatusInternalServerError && errors.As(err, &appError) {
		public.Code, public.Message = appError.Code, appError.Message
	}

	return &Response{Status: status, Payload: public, InternalErr: err}
}

func errorType(status int) string {
	return "srn:error:" + strings.ReplaceAll(strings.ToLower(http.StatusText(status)), " ", "-")
}
