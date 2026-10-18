package report

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
)

func TestRepeatedToolCallsReuseTheSameSnapshot(t *testing.T) {
	t.Parallel()

	id := "00000000-0000-4000-8000-000000000001"
	data := NewMockData(gomock.NewController(t))
	snapshot := demand.Detail{Demand: entity.Demand{ID: id, Title: "Contexto inicial"}}
	data.EXPECT().Detail(gomock.Any(), id).Return(snapshot, nil).Times(1)
	tools := NewTools(data, Request{})

	first, err := tools.Call(t.Context(), "get_demand_context", json.RawMessage(`{"id":"`+id+`"}`))
	if err != nil {
		t.Fatal(err)
	}

	second, err := tools.Call(t.Context(), "get_demand_context", json.RawMessage(`{ "id": "`+id+`" }`))
	if err != nil {
		t.Fatal(err)
	}

	firstDetail, firstOK := first.(demand.Detail)
	secondDetail, secondOK := second.(demand.Detail)

	if !firstOK || !secondOK {
		t.Fatal("expected demand details")
	}

	if len(tools.Snapshots) != 1 || firstDetail.Demand.Title != "Contexto inicial" || secondDetail.Demand.Title != "Contexto inicial" {
		t.Fatal("repeated query discarded the original snapshot")
	}
}

func TestToolScopeCannotBeExpandedByModel(t *testing.T) {
	t.Parallel()

	old := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	recent := old.AddDate(0, 1, 0)
	data := NewMockData(gomock.NewController(t))
	rows := []entity.Demand{{ID: "00000000-0000-4000-8000-000000000001", Status: entity.Done, CreatedAt: old, CompletedAt: &recent}, {ID: "00000000-0000-4000-8000-000000000002", CreatedAt: old}}
	data.EXPECT().List(gomock.Any(), "", "", "").Return(rows, nil).Times(1)
	data.EXPECT().Detail(gomock.Any(), rows[1].ID).Return(demand.Detail{Demand: rows[1]}, nil).Times(1)
	tools := NewTools(data, Request{From: "2026-10-01", To: "2026-10-31"})

	result, err := tools.Call(t.Context(), "list_demands", json.RawMessage(`{"status":""}`))
	if err != nil {
		t.Fatal(err)
	}

	payload, ok := result.(map[string]any)
	if !ok {
		t.Fatal("expected tool payload")
	}

	visibleRows, ok := payload["demands"].([]entity.Demand)
	if !ok {
		t.Fatal("expected demand rows")
	}

	if len(visibleRows) != 1 {
		t.Fatalf("scope: %+v", visibleRows)
	}

	_, err = tools.Call(t.Context(), "get_demand_context", json.RawMessage(`{"id":"00000000-0000-4000-8000-000000000002"}`))
	if !errors.Is(err, erring.ErrNotFound) {
		t.Fatalf("escaped scope: %v", err)
	}

	_, err = tools.Call(t.Context(), "execute_sql", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("accepted unknown tool")
	}

	content := Content{Title: "Relatório", Summary: "Resumo", SourceIDs: []string{"invented"}}
	if ValidateContent(content, tools.Sources) == nil {
		t.Fatal("accepted invented source")
	}

	content.SourceIDs = nil
	content.Tables = []Table{{Columns: []string{"A", "B"}, Rows: [][]string{{"A"}}}}

	if ValidateContent(content, tools.Sources) == nil {
		t.Fatal("accepted malformed table")
	}
}
