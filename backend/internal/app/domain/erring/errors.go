package erring

import (
	"fmt"
)

var (
	ErrInvalidInput       = NewAppError("request:invalid", "Os dados informados são inválidos.")
	ErrInvalidTransition  = NewAppError("demand:invalid-transition", "A demanda não pode ser movida para essa etapa.")
	ErrConflict           = NewAppError("resource:conflict", "A operação conflita com o estado atual do recurso.")
	ErrNotFound           = NewAppError("resource:not-found", "Recurso não encontrado.")
	ErrForbidden          = NewAppError("request:forbidden", "Você não tem permissão para realizar esta ação.")
	ErrTooManyRequests    = NewAppError("request:too-many-requests", "Muitas tentativas. Aguarde alguns minutos.")
	ErrPayloadTooLarge    = NewAppError("request:payload-too-large", "O conteúdo enviado excede o limite permitido.")
	ErrMethodNotAllowed   = NewAppError("request:method-not-allowed", "O método HTTP não é permitido para este recurso.")
	ErrServiceUnavailable = NewAppError("service:unavailable", "O serviço está temporariamente indisponível.")
)

// Wrap adds operation context while preserving nil and errors.Is/errors.As.
func Wrap(err error, operation string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", operation, err)
}
