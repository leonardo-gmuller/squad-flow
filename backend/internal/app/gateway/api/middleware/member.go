package middleware

import (
	"net/http"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/auth"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/notification"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/response"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
)

func RequireMember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := ctxkey.GetUser(r.Context())
		if !ok {
			response.SendContext(r.Context(), w, response.AppError(auth.ErrUnauthorized))

			return
		}

		if err := notification.RequireMember(user); err != nil {
			response.SendContext(r.Context(), w, response.AppError(err))

			return
		}

		next.ServeHTTP(w, r)
	})
}
