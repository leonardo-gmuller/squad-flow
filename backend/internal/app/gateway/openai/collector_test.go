package openai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
)

func newCollectionTools(t *testing.T, calls int) *MockToolset {
	t.Helper()
	tools := NewMockToolset(gomock.NewController(t))
	tools.EXPECT().SourceIDs().Return([]string{}).AnyTimes()
	tools.EXPECT().Definitions().Return([]report.Tool{{Name: "get_demand_context", Description: "Histórico da demanda", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}}, "required": []string{"id"}, "additionalProperties": false}}}).AnyTimes()
	tools.EXPECT().Call(gomock.Any(), "get_demand_context", gomock.Any()).Return(map[string]string{"evidence": "Histórico consultado"}, nil).Times(calls)

	return tools
}

func toolOutput(id, args string) map[string]string {
	return map[string]string{"type": "function_call", "name": "get_demand_context", "call_id": id, "arguments": args}
}

func writeCollectionResponse(t *testing.T, w http.ResponseWriter, output []any) {
	t.Helper()
	writeTestJSON(t, w, map[string]any{"status": "completed", "output": output})
}

func writeValidReport(t *testing.T, w http.ResponseWriter) {
	t.Helper()

	content := report.Content{Title: "Relatório de demandas", Summary: "Análise baseada nos dados consultados", Metrics: []report.Metric{}, Sections: []report.Section{}, Charts: []report.Chart{}, Tables: []report.Table{}, Recommendations: []string{}, MissingData: []string{}, SourceIDs: []string{}}
	encoded := marshalTestJSON(t, content)
	writeCollectionResponse(t, w, []any{map[string]any{"type": "message", "content": []any{map[string]string{"type": "output_text", "text": string(encoded)}}}})
}

func TestRepeatedQueriesFinishCollectionAndGenerateReport(t *testing.T) {
	t.Parallel()

	collectorCalls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)

			return
		}

		if body["tools"] != nil {
			collectorCalls++

			args := `{"id":"00000000-0000-4000-8000-000000000001"}`
			if collectorCalls == 2 {
				args = `{ "id": "00000000-0000-4000-8000-000000000001" }`
			}

			writeCollectionResponse(t, w, []any{toolOutput(strconv.Itoa(collectorCalls), args)})

			return
		}

		if body["tool_choice"] != "none" {
			t.Error("analysis allowed new tool calls")
		}

		writeValidReport(t, w)
	}))
	defer server.Close()

	client := New("fake-key", "test-model")
	client.Endpoint = server.URL
	tools := newCollectionTools(t, 1)

	out, err := client.Generate(t.Context(), "Avalie o histórico das demandas", tools, func(string, string) {})
	if err != nil {
		t.Fatal(err)
	}

	if collectorCalls != 2 || len(out.MissingData) != 1 {
		t.Fatalf("unbounded/repeated collection: rounds=%d limitations=%v", collectorCalls, out.MissingData)
	}
}

func TestUniqueCallBudgetReturnsReportAndAnswersEveryPendingCall(t *testing.T) {
	t.Parallel()

	collectorCalls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)

			return
		}

		if body["tools"] != nil {
			collectorCalls++

			if body["parallel_tool_calls"] != true {
				t.Error("collector does not allow batches")
			}

			output := []any{}

			for i := range 30 {
				args := fmt.Sprintf(`{"id":"00000000-0000-4000-8000-%012d"}`, i)
				output = append(output, toolOutput(strconv.Itoa(i), args))
			}

			writeCollectionResponse(t, w, output)

			return
		}

		if body["tool_choice"] != "none" {
			t.Error("analysis allowed new tool calls")
		}

		input, _ := json.Marshal(body["input"])
		if len(input) == 0 {
			t.Error("collected evidence missing")
		}

		writeValidReport(t, w)
	}))
	defer server.Close()

	client := New("fake-key", "test-model")
	client.Endpoint = server.URL
	tools := newCollectionTools(t, 48)

	collected, err := client.collect(t.Context(), "Avalie as demandas", tools)
	if err != nil {
		t.Fatal(err)
	}

	results := 0

	for _, item := range collected.Input {
		if m, ok := item.(map[string]any); ok && m["type"] == "function_call_output" {
			results++
		}
	}

	if collectorCalls != 1 || results != 30 || len(collected.Limitations) != 1 {
		t.Fatalf("invalid budget handoff: rounds=%d outputs=%d limitations=%v", collectorCalls, results, collected.Limitations)
	}

	out, err := client.Generate(t.Context(), "Avalie as demandas", tools, func(string, string) {})
	if err != nil {
		t.Fatal(err)
	}

	if out.Title == "" || len(out.MissingData) != 1 {
		t.Fatal("limited report was not generated transparently")
	}
}

func TestEightRoundsHandOffCollectedEvidenceInsteadOfFailing(t *testing.T) {
	t.Parallel()

	collectorCalls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)

			return
		}

		if body["tools"] != nil {
			collectorCalls++
			args := fmt.Sprintf(`{"id":"00000000-0000-4000-8000-%012d"}`, collectorCalls)
			writeCollectionResponse(t, w, []any{toolOutput(strconv.Itoa(collectorCalls), args)})

			return
		}

		writeValidReport(t, w)
	}))
	defer server.Close()

	client := New("fake-key", "test-model")
	client.Endpoint = server.URL
	tools := newCollectionTools(t, 8)

	out, err := client.Generate(t.Context(), "Avalie todas as demandas", tools, func(string, string) {})
	if err != nil {
		t.Fatal(err)
	}

	if collectorCalls != 8 || len(out.MissingData) != 1 {
		t.Fatalf("round limit did not hand off: rounds=%d limitations=%v", collectorCalls, out.MissingData)
	}
}
