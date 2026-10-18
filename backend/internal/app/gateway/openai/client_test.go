package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
)

func TestResponseStorageIsConfigurableAndCarriesJobMetadata(t *testing.T) {
	t.Parallel()

	for _, enabled := range []bool{false, true} {
		t.Run(strconv.FormatBool(enabled), func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					Store    bool              `json:"store"`
					Metadata map[string]string `json:"metadata"`
				}

				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}

				if body.Store != enabled || body.Metadata["agent"] != "análise" || body.Metadata["report_job_id"] != "job-test" || body.Metadata["application"] != "squadflow" {
					t.Errorf("incorrect storage/metadata: %+v", body)
				}

				writeTestJSON(t, w, map[string]any{"status": "completed", "output": []any{}})
			}))
			defer server.Close()

			client := New("fake", "test")
			client.StoreResponses = enabled
			client.Endpoint = server.URL
			ctx := ctxkey.PutAgent(ctxkey.PutRequestID(t.Context(), "job-test"), "análise")

			if _, err := client.request(ctx, map[string]any{}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRequestRespectsContextCancellation(t *testing.T) {
	t.Parallel()

	for _, phase := range []string{"headers", "body"} {
		t.Run(phase, func(t *testing.T) {
			t.Parallel()

			started := make(chan struct{})

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if _, err := io.Copy(io.Discard, r.Body); err != nil {
					t.Error(err)

					return
				}

				if phase == "body" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)

					if err := http.NewResponseController(w).Flush(); err != nil {
						t.Error(err)

						return
					}
				}

				close(started)
				<-r.Context().Done()
			}))
			defer server.Close()

			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			client := New("fake-key", "test-model")
			client.Endpoint = server.URL
			result := make(chan error, 1)

			go func() {
				_, err := client.request(ctx, map[string]any{"model": "test-model"})
				result <- err
			}()

			select {
			case <-started:
			case <-ctx.Done():
				t.Fatal("request did not reach provider")
			}

			cancel()

			select {
			case err := <-result:
				if !errors.Is(err, context.Canceled) || !errors.Is(err, ErrProvider) {
					t.Fatalf("context cancellation was not preserved: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("request did not stop after context cancellation")
			}
		})
	}
}

func newMetricsTools(t *testing.T, calls int) *MockToolset {
	t.Helper()
	tools := NewMockToolset(gomock.NewController(t))
	tools.EXPECT().SourceIDs().Return([]string{}).AnyTimes()
	tools.EXPECT().Definitions().Return([]report.Tool{{Name: "get_metrics", Description: "Indicadores", InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "required": []string{}, "additionalProperties": false}}}).AnyTimes()
	tools.EXPECT().Call(gomock.Any(), "get_metrics", json.RawMessage(`{}`)).Return(map[string]int{"delivered": 3}, nil).Times(calls)

	return tools
}

func TestThreeAgentsUseToolsAndStrictStructuredOutputs(t *testing.T) {
	t.Parallel()

	calls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++

		var input map[string]any
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Error(err)
		}

		if input["store"] != false || r.Header.Get("Authorization") != "Bearer fake-key" {
			t.Error("storage/authentication")
		}

		w.Header().Set("Content-Type", "application/json")

		if calls == 1 {
			writeTestJSON(t, w, map[string]any{"status": "completed", "output": []any{map[string]string{"type": "function_call", "name": "get_metrics", "call_id": "call-1", "arguments": "{}"}}})

			return
		}

		text := "Dados coletados"

		if calls == 2 {
			encoded := marshalTestJSON(t, input["input"])
			if !strings.Contains(string(encoded), "function_call_output") {
				t.Error("missing tool output")
			}
		}

		if calls >= 3 {
			textOptions, ok := input["text"].(map[string]any)
			if !ok {
				t.Error("missing text format")

				return
			}

			format, ok := textOptions["format"].(map[string]any)
			if !ok {
				t.Error("missing structured format")

				return
			}

			if format["strict"] != true || format["type"] != "json_schema" {
				t.Error("schema is not strict")
			}

			text = `{"title":"Entregas","summary":"3 entregas","html":"<main><h1>Entregas</h1><svg><text>3</text></svg></main>","metrics":[{"label":"Entregas","value":3,"unit":"demandas"}],"sections":[],"charts":[],"tables":[],"recommendations":[],"missing_data":[],"source_ids":[]}`
		}

		writeTestJSON(t, w, map[string]any{"status": "completed", "output": []any{map[string]any{"type": "message", "content": []any{map[string]string{"type": "output_text", "text": text}}}}})
	}))
	defer server.Close()

	client := New("fake-key", "test-model")
	client.Endpoint = server.URL
	tools := newMetricsTools(t, 1)
	agents := []string{}

	out, err := client.Generate(t.Context(), "Analise entregas", tools, func(agent, _ string) { agents = append(agents, agent) })
	if err != nil {
		t.Fatal(err)
	}

	if calls != 4 || out.Metrics[0].Value != 3 || len(agents) != 3 || !strings.Contains(out.HTML, "<svg>") {
		t.Fatalf("agents: calls=%d output=%+v agents=%v", calls, out, agents)
	}
}

func TestIncompleteAndRefusedResponsesAreRejected(t *testing.T) {
	t.Parallel()

	for _, body := range []string{`{"status":"incomplete","output":[]}`, `{"status":"completed","output":[{"type":"message","content":[{"type":"refusal"}]}]}`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if _, err := w.Write([]byte(body)); err != nil {
				t.Error(err)
			}
		}))
		client := New("fake", "test")
		client.Endpoint = server.URL

		if _, err := client.structured(t.Context(), "test"); !errors.Is(err, ErrProvider) {
			t.Fatalf("accepted invalid response: %v", err)
		}

		server.Close()
	}
}

func TestEmptyReviewerTitleIsRepairedBeforeReturning(t *testing.T) {
	t.Parallel()

	calls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		text := "Dados coletados"

		if calls > 1 {
			title := "Relatório do workspace"
			if calls == 3 {
				title = ""
			}

			content := report.Content{Title: title, Summary: "Três demandas consultadas", Metrics: []report.Metric{}, Sections: []report.Section{}, Charts: []report.Chart{}, Tables: []report.Table{}, Recommendations: []string{}, MissingData: []string{}, SourceIDs: []string{}}
			encoded, _ := json.Marshal(content)
			text = string(encoded)
		}

		writeTestJSON(t, w, map[string]any{"status": "completed", "output": []any{map[string]any{"type": "message", "content": []any{map[string]string{"type": "output_text", "text": text}}}}})
	}))
	defer server.Close()

	client := New("fake-key", "test-model")
	client.Endpoint = server.URL

	out, err := client.Generate(t.Context(), "Analise o workspace", newMetricsTools(t, 0), func(string, string) {})
	if err != nil {
		t.Fatal(err)
	}

	if calls != 4 || out.Title != "Relatório do workspace" {
		t.Fatalf("missing repair: calls=%d title=%q", calls, out.Title)
	}
}

//nolint:paralleltest // Captures the process-wide logger; parallel tests must stay suspended.
func TestQuotaErrorIsActionableAndDoesNotLogSecrets(t *testing.T) {
	var logs bytes.Buffer

	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))

	defer slog.SetDefault(previous)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("x-request-id", "req-quota-test")
		w.WriteHeader(http.StatusTooManyRequests)

		if _, err := w.Write([]byte(`{"error":{"code":"insufficient_quota","message":"sensitive-provider-message"}}`)); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	client := New("secret-key-for-test", "test-model")
	client.Endpoint = server.URL

	_, err := client.structured(t.Context(), "private-prompt-for-test")
	if !errors.Is(err, ErrProvider) || !strings.Contains(err.Error(), "saldo") {
		t.Fatalf("wrong error: %v", err)
	}

	for _, secret := range []string{"secret-key-for-test", "private-prompt-for-test", "sensitive-provider-message"} {
		if strings.Contains(logs.String(), secret) {
			t.Fatalf("log leaked %s", secret)
		}
	}

	if !strings.Contains(logs.String(), "req-quota-test") {
		t.Fatal("missing provider request ID")
	}
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()

	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Error(err)
	}
}

func marshalTestJSON(t *testing.T, value any) []byte {
	t.Helper()

	result, err := json.Marshal(value)
	if err != nil {
		t.Errorf("marshal fixture: %v", err)
	}

	return result
}
