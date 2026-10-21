// Package response builds and writes the shared HTTP success and error contracts.
package response

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

// Handle keeps response serialization out of application handlers.
func Handle(handler func(*http.Request) *Response) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		SendContext(r.Context(), w, handler(r))
	}
}

func Send(w http.ResponseWriter, result *Response) {
	SendContext(context.Background(), w, result)
}

// SendContext correlates error logs with the request span without changing the response contract.
func SendContext(ctx context.Context, w http.ResponseWriter, result *Response) {
	if result == nil {
		result = AppError(nil)
	}

	if result.Status >= http.StatusInternalServerError {
		slog.ErrorContext(ctx, "request failed", "status", result.Status, "error", result.InternalErr)
	}

	for key, values := range result.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	for _, cookie := range result.Cookies {
		http.SetCookie(w, cookie)
	}

	JSON(w, result.Status, result.Payload)
}

// JSON also supports protocol-specific payloads such as MCP JSON-RPC responses.
func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("X-Content-Type-Options", "nosniff")

	if value == nil || status == http.StatusNoContent {
		w.WriteHeader(status)

		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}
