package middleware

import (
	"net/http"
	"strings"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/response"
)

type OriginPolicy struct {
	allowed map[string]struct{}
}

func NewOriginPolicy(origins string) *OriginPolicy {
	policy := &OriginPolicy{allowed: make(map[string]struct{})}

	for _, origin := range strings.Split(origins, ",") {
		if value := strings.TrimSpace(origin); value != "" {
			policy.allowed[value] = struct{}{}
		}
	}

	return policy
}

func (p *OriginPolicy) Allowed(origin string) bool {
	_, ok := p.allowed[origin]

	return ok
}

// Protect requires an allowed Origin when present and the existing JSON/client
// headers for mutations, including Bearer clients that omit Origin.
func (p *OriginPolicy) Protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" && !p.Allowed(origin) {
			response.Send(w, response.AppError(erring.ErrOriginNotAllowed))

			return
		}

		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if r.Header.Get("X-Squadflow-Client") != "web" || !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
				response.Send(w, response.AppError(erring.ErrCSRFProtection))

				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
