package erring

var (
	ErrUnauthorized           = NewAppError("auth:unauthorized", "Sua sessão é inválida ou expirou. Entre novamente.")
	ErrAdminRequired          = NewAppError("auth:admin-required", "Acesso permitido somente ao administrador.").WithCause(ErrForbidden)
	ErrInvalidCredentials     = NewAppError("auth:invalid-credentials", "Email ou senha inválidos.").WithCause(ErrUnauthorized)
	ErrEmailAlreadyUsed       = NewAppError("auth:email-already-used", "Este email já está cadastrado.").WithCause(ErrConflict)
	ErrInvalidResetToken      = NewAppError("auth:invalid-reset-token", "O link de redefinição é inválido ou expirou. Solicite outro ao administrador.").WithCause(ErrInvalidInput)
	ErrInvalidCurrentPassword = NewAppError("auth:invalid-current-password", "A senha atual está incorreta.").WithCause(ErrInvalidInput)
	ErrOriginNotAllowed       = NewAppError("auth:origin-not-allowed", "Origem não permitida.").WithCause(ErrForbidden)
	ErrCSRFProtection         = NewAppError("auth:csrf-protection", "Requisição sem proteção CSRF.").WithCause(ErrForbidden)
)
