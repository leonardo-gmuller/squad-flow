package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
)

const (
	maxCollectionRounds = 8
	maxUniqueToolCalls  = 24
)

type collection struct {
	Input       []any
	Limitations []string
}

func toolCallKey(name, args string) string {
	var parameters any
	if json.Unmarshal([]byte(args), &parameters) == nil {
		if canonical, err := json.Marshal(parameters); err == nil {
			return name + ":" + string(canonical)
		}
	}

	return name + ":" + args
}

func (c *Client) collect(ctx context.Context, prompt string, registry report.Toolset) (collection, error) {
	tools := []any{}
	for _, tool := range registry.Definitions() {
		tools = append(tools, map[string]any{"type": "function", "name": tool.Name, "description": tool.Description, "parameters": tool.InputSchema, "strict": true})
	}

	result := collection{Input: []any{message("user", prompt+"\nColete somente os dados adicionais necessários. Os dados já fornecidos e os resultados de ferramentas permanecem válidos durante este relatório: reutilize-os e não repita consultas. Quando precisar do histórico ou dos testes de várias demandas, solicite seus contextos no mesmo lote usando os IDs reais. Há até 24 consultas únicas. Quando tiver evidências suficientes, encerre a coleta com um resumo e indique as limitações.")}}
	cache := map[string]json.RawMessage{}
	uniqueCalls := 0
	finish := func(reason, limitation string) collection {
		slog.Info("report collection completed", "reason", reason, "unique_tool_calls", uniqueCalls)

		if limitation != "" {
			result.Limitations = append(result.Limitations, limitation)
			result.Input = append(result.Input, message("user", "Coleta encerrada pelo backend. Registre esta limitação e analise somente dados efetivamente consultados: "+limitation))
		}

		return result
	}

	for range maxCollectionRounds {
		response, err := c.request(ctxkey.PutAgent(ctx, "coleta"), map[string]any{"model": c.Model, "instructions": instructions, "input": result.Input, "tools": tools, "tool_choice": "auto", "parallel_tool_calls": true, "max_output_tokens": 4000, "include": []string{"reasoning.encrypted_content"}})
		if err != nil {
			return result, err
		}

		round, roundErr := collectRound(ctx, registry, response.Output, cache, uniqueCalls)
		if roundErr != nil {
			return result, roundErr
		}

		result.Input = append(result.Input, round.Input...)
		uniqueCalls += round.NewCalls

		if round.Calls == 0 {
			if _, err = outputText(response); err != nil {
				return result, err
			}

			return finish("model_completed", ""), nil
		}

		if round.BudgetReached || uniqueCalls >= maxUniqueToolCalls {
			return finish("tool_budget", "A coleta foi limitada a 24 consultas únicas. Demandas ou evidências adicionais não consultadas não estão cobertas por esta análise."), nil
		}

		if round.NewCalls == 0 {
			return finish("repeated_queries", "O agente repetiu consultas já respondidas. A coleta foi encerrada e a análise considera somente as evidências obtidas."), nil
		}
	}

	return finish("round_budget", "A coleta foi encerrada após oito rodadas. A análise considera somente as evidências obtidas; contextos adicionais não consultados permanecem ausentes."), nil
}

func appendLimitations(content report.Content, limitations []string) report.Content {
	for _, limitation := range limitations {
		found := false

		for _, existing := range content.MissingData {
			if existing == limitation {
				found = true

				break
			}
		}

		if !found {
			content.MissingData = append(content.MissingData, limitation)
		}
	}

	return content
}

type collectionRound struct {
	Input           []any
	Calls, NewCalls int
	BudgetReached   bool
}

func collectRound(ctx context.Context, registry report.Toolset, output []json.RawMessage, cache map[string]json.RawMessage, priorCalls int) (collectionRound, error) {
	result := collectionRound{}
	for _, raw := range output {
		result.Input = append(result.Input, raw)

		var item outputItem

		if err := json.Unmarshal(raw, &item); err != nil {
			return result, ErrProvider
		}

		if item.Type != "function_call" {
			continue
		}

		if item.CallID == "" {
			return result, fmt.Errorf("%w: chamada de ferramenta sem identificador", ErrProvider)
		}

		result.Calls++
		key := toolCallKey(item.Name, item.Arguments)

		encoded, cached := cache[key]
		if !cached {
			limited := priorCalls+result.NewCalls >= maxUniqueToolCalls
			if limited {
				result.BudgetReached = true
			} else {
				result.NewCalls++
			}

			var err error

			encoded, err = collectToolResult(ctx, registry, item, limited)
			if err != nil {
				return result, err
			}

			cache[key] = encoded
		}

		result.Input = append(result.Input, map[string]any{"type": "function_call_output", "call_id": item.CallID, "output": string(encoded)})
	}

	return result, nil
}

func collectToolResult(ctx context.Context, registry report.Toolset, item outputItem, limited bool) (json.RawMessage, error) {
	var data any
	if limited {
		data = map[string]string{"error": "consulta não executada: orçamento de coleta atingido; registre o dado ausente no relatório"}
	} else {
		var err error

		data, err = registry.Call(ctx, item.Name, json.RawMessage(item.Arguments))
		if err != nil {
			detail := "não foi possível consultar os dados"
			if errors.Is(err, erring.ErrInvalidInput) || errors.Is(err, erring.ErrNotFound) {
				detail = err.Error()
			}

			data = map[string]string{"error": detail}
		}
	}

	encoded, err := json.Marshal(data)

	return encoded, erring.Wrap(err, "openai.collectToolResult")
}
