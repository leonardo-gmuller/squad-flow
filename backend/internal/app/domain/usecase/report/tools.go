package report

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

// Tools is the same read-only registry used by the agent and the MCP adapter.
// Request scope is enforced here, independently of model tool arguments.
type Tools struct {
	Data      Data
	Scope     Request
	Sources   map[string]entity.Demand
	Snapshots map[string]any
}

func NewTools(data Data, scope Request) *Tools {
	return &Tools{Data: data, Scope: scope, Sources: map[string]entity.Demand{}, Snapshots: map[string]any{}}
}

func (t *Tools) SourceIDs() []string {
	ids := make([]string, 0, len(t.Sources))
	for id := range t.Sources {
		ids = append(ids, id)
	}

	sort.Strings(ids)

	return ids
}

func object(properties map[string]any, required ...string) map[string]any {
	if required == nil {
		required = []string{}
	}

	return map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
}

func (t *Tools) Definitions() []Tool {
	contextID := map[string]any{"type": "string", "description": "UUID real da demanda, obtido em list_demands ou nos dados já fornecidos."}
	if ids := t.SourceIDs(); len(ids) > 0 && len(ids) <= 250 {
		contextID["enum"] = ids
	}

	return []Tool{
		{"list_demands", "Lista demandas do workspace dentro do escopo solicitado, com contexto e datas. Máximo 2000; informa truncamento.", object(map[string]any{"status": map[string]any{"type": "string", "enum": []string{"", "backlog", "in_progress", "in_qa", "changes_requested", "done"}}}, "status")},
		{"get_metrics", "Calcula indicadores determinísticos sobre todas as demandas do escopo. Datas filtram a conclusão para finalizadas e criação para demais.", object(map[string]any{})},
		{"get_demand_context", "Obtém a demanda, todos os testes e o histórico de movimentações. Use para avaliar causas, QA, devoluções e tempo por etapa. Reutilize contextos já obtidos e peça múltiplas demandas no mesmo lote.", object(map[string]any{"id": contextID}, "id")},
	}
}

func (t *Tools) inScope(item entity.Demand) bool {
	if t.Scope.DemandID != "" && t.Scope.DemandID != item.ID {
		return false
	}

	date := item.CreatedAt
	if item.Status == entity.Done && item.CompletedAt != nil {
		date = *item.CompletedAt
	}

	day := date.UTC().Format("2006-01-02")

	return (t.Scope.From == "" || day >= t.Scope.From) && (t.Scope.To == "" || day <= t.Scope.To)
}

func (t *Tools) rows(ctx context.Context) ([]entity.Demand, error) {
	rows, err := t.Data.List(ctx, "", "", "")
	if err != nil {
		return nil, erring.Wrap(err, "report.rows")
	}

	out := []entity.Demand{}

	for _, d := range rows {
		if t.inScope(d) {
			out = append(out, d)
		}
	}

	return out, nil
}

func (t *Tools) Call(ctx context.Context, name string, args json.RawMessage) (any, error) {
	var input struct {
		Status string `json:"status"`
		ID     string `json:"id"`
	}

	dec := json.NewDecoder(strings.NewReader(string(args)))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&input); err != nil {
		return nil, erring.ErrInvalidInput.WithMessage("argumentos da ferramenta inválidos")
	}

	var parameters map[string]json.RawMessage
	if err := json.Unmarshal(args, &parameters); err != nil || parameters == nil {
		return nil, erring.ErrInvalidInput.WithMessage("argumentos devem ser um objeto JSON")
	}

	canonical, err := json.Marshal(parameters)
	if err != nil {
		return nil, erring.Wrap(err, "report.Call")
	}

	key := name + ":" + string(canonical)
	if cached, ok := t.Snapshots[key]; ok {
		return cached, nil
	}

	var out any

	switch name {
	case "list_demands":
		out, err = t.listDemands(ctx, input.Status)
		if err != nil {
			return nil, err
		}
	case "get_metrics":
		rows, queryErr := t.rows(ctx)
		if queryErr != nil {
			return nil, queryErr
		}

		out = demand.CalculateMetrics(rows, time.Now().UTC())
	case "get_demand_context":
		if !identifier.IsUUID(input.ID) {
			return nil, erring.ErrInvalidInput
		}

		if t.Scope.DemandID != "" && input.ID != t.Scope.DemandID {
			return nil, erring.ErrNotFound
		}

		detail, detailErr := t.Data.Detail(ctx, input.ID)
		if detailErr != nil {
			return nil, erring.Wrap(detailErr, "report.Call")
		}

		if !t.inScope(detail.Demand) {
			return nil, erring.ErrNotFound
		}

		t.Sources[input.ID] = detail.Demand
		out = detail
	default:
		return nil, erring.ErrInvalidInput.WithMessage("ferramenta desconhecida")
	}

	b, err := json.Marshal(out)
	if err != nil {
		return nil, erring.Wrap(err, "report.Call")
	}

	if len(b) > 800000 {
		return nil, erring.ErrInvalidInput.WithMessage("contexto muito grande; restrinja período ou demanda")
	}

	t.Snapshots[key] = out

	return out, nil
}

func (t *Tools) listDemands(ctx context.Context, status string) (any, error) {
	if status != "" && !entity.IsStatus(entity.Status(status)) {
		return nil, erring.ErrInvalidInput
	}

	rows, err := t.rows(ctx)
	if err != nil {
		return nil, err
	}

	filtered := []entity.Demand{}

	for _, item := range rows {
		if status == "" || string(item.Status) == status {
			filtered = append(filtered, item)
		}
	}

	total := len(filtered)
	if total > 2000 {
		filtered = filtered[:2000]
	}

	for _, item := range filtered {
		t.Sources[item.ID] = item
	}

	return map[string]any{"demands": filtered, "total": total, "truncated": total > len(filtered), "date_semantics": "conclusão para finalizadas; criação para demais; dias UTC"}, nil
}
