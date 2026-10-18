package erring

// AppError carries a stable code and a public message, independent of transport.
type AppError struct {
	Code    string
	Message string
	cause   error
}

func NewAppError(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func (e *AppError) Error() string { return e.Message }

func (e *AppError) Unwrap() error { return e.cause }

// WithMessage preserves the error category without modifying a shared sentinel.
func (e *AppError) WithMessage(message string) *AppError {
	return &AppError{Code: e.Code, Message: message, cause: e}
}

func (e *AppError) WithCause(cause error) *AppError {
	return &AppError{Code: e.Code, Message: e.Message, cause: cause}
}
