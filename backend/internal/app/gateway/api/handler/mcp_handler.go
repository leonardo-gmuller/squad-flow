package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/api/response"
)

// MCP implements stateless Streamable HTTP with JSON responses, no SSE session.
// Authentication and Origin/CSRF validation run before this adapter.
type MCP struct{ Data report.Data }

func (h MCP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)

		return
	}

	var req struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
	}

	if !decode(w, r, &req) {
		return
	}

	reply := func(result any) {
		response.JSON(w, 200, map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}

	fault := func(code int, message string) {
		response.JSON(w, 200, map[string]any{"jsonrpc": "2.0", "id": req.ID, "error": map[string]any{"code": code, "message": message}})
	}
	if req.JSONRPC != "2.0" {
		fault(-32600, "invalid JSON-RPC request")

		return
	}

	if len(req.ID) == 0 {
		if req.Method == "notifications/initialized" || req.Method == "notifications/cancelled" {
			w.WriteHeader(http.StatusAccepted)

			return
		}

		fault(-32600, "id required")

		return
	}

	tools := report.NewTools(h.Data, report.Request{})

	switch req.Method {
	case "initialize":
		var input struct {
			ProtocolVersion string `json:"protocolVersion"`
		}

		if json.Unmarshal(req.Params, &input) != nil {
			fault(-32602, "invalid initialization")

			return
		}

		version := input.ProtocolVersion
		if version != "2025-03-26" && version != "2025-06-18" && version != "2025-11-25" {
			version = "2025-06-18"
		}

		reply(map[string]any{"protocolVersion": version, "capabilities": map[string]any{"tools": map[string]any{"listChanged": false}}, "serverInfo": map[string]string{"name": "squadflow-reports", "version": "0.5.0"}})
	case "ping":
		reply(map[string]any{})
	case "tools/list":
		reply(map[string]any{"tools": tools.Definitions()})
	case "tools/call":
		var input struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}

		if json.Unmarshal(req.Params, &input) != nil {
			fault(-32602, "invalid tool arguments")

			return
		}

		out, err := tools.Call(r.Context(), input.Name, input.Arguments)
		if err != nil {
			message := "não foi possível consultar os dados"
			if errors.Is(err, erring.ErrInvalidInput) || errors.Is(err, erring.ErrNotFound) {
				message = err.Error()
			}

			reply(map[string]any{"isError": true, "content": []any{map[string]string{"type": "text", "text": message}}})

			return
		}

		text, err := json.Marshal(out)
		if err != nil {
			fault(-32603, "serialization failed")

			return
		}

		reply(map[string]any{"content": []any{map[string]string{"type": "text", "text": string(text)}}, "structuredContent": out, "isError": false})
	default:
		fault(-32601, "method not found")
	}
}
