package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/auth"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/response"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
)

type Authenticator interface {
	Authenticate(ctx context.Context, raw string) (auth.Principal, error)
}

// AccessToken gives a Bearer credential precedence over the access cookie.
func AccessToken(r *http.Request) string {
	value, ok := ctxkey.GetAuthorizationHeader(r.Context())
	if !ok {
		value = r.Header.Get("Authorization")
	}

	if strings.HasPrefix(value, "Bearer ") {
		return strings.TrimPrefix(value, "Bearer ")
	}

	if cookie, err := r.Cookie("sf_access"); err == nil {
		return cookie.Value
	}

	return ""
}

func RequireAuth(service Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, err := service.Authenticate(r.Context(), AccessToken(r))
			if err != nil {
				response.SendContext(r.Context(), w, response.AppError(err))

				return
			}

			ctx := ctxkey.PutUser(r.Context(), principal.User)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
