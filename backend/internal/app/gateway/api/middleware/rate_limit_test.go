package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/middleware"
)

func TestRateLimiterUsesSourceAddressAndSetsRetryAfter(t *testing.T) {
	t.Parallel()

	limiter := middleware.NewRateLimiter(2, time.Minute)
	handler := limiter.Limit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))

	for attempt := range 3 {
		request := httptest.NewRequest(http.MethodPost, "/password-reset", nil)
		request.RemoteAddr = "127.0.0.1:1234"
		request.Header.Set("X-Forwarded-For", "different-forged-address")

		recorded := httptest.NewRecorder()
		handler.ServeHTTP(recorded, request)

		if attempt < 2 && recorded.Code != http.StatusNoContent {
			t.Fatal("valid attempt was throttled")
		}

		if attempt == 2 && (recorded.Code != http.StatusTooManyRequests || recorded.Header().Get("Retry-After") == "") {
			t.Fatal("exhausted limit did not return retry information")
		}
	}

	request := httptest.NewRequest(http.MethodPost, "/password-reset", nil)
	request.RemoteAddr = "127.0.0.2:1234"
	recorded := httptest.NewRecorder()
	handler.ServeHTTP(recorded, request)

	if recorded.Code != http.StatusNoContent {
		t.Fatal("independent source was throttled")
	}
}
