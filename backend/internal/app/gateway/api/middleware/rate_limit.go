package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/response"
)

// RateLimiter uses the socket's source address, without trusting proxy headers.
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]rateEntry
	limit   int
	window  time.Duration
}

type rateEntry struct {
	count   int
	expires time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{entries: make(map[string]rateEntry), limit: limit, window: window}
}

func (l *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}

		now := time.Now()

		l.mu.Lock()

		for key, entry := range l.entries {
			if !entry.expires.After(now) {
				delete(l.entries, key)
			}
		}

		entry := l.entries[host]
		if entry.count == 0 {
			entry.expires = now.Add(l.window)
		}

		entry.count++

		allowed := entry.count <= l.limit && (len(l.entries) < 10000 || l.entries[host].count > 0)
		if allowed {
			l.entries[host] = entry
		}

		l.mu.Unlock()

		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(max(1, int(time.Until(entry.expires).Seconds())+1)))
			response.SendContext(r.Context(), w, response.AppError(erring.ErrTooManyRequests))

			return
		}

		next.ServeHTTP(w, r)
	})
}
