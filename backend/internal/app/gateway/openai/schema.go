package openai

import "sort"

// Schema carries a standalone HTML document plus metadata and optional legacy components.
func Schema(sourceIDs ...string) map[string]any {
	str := map[string]any{"type": "string"}
	num := map[string]any{"type": "number"}
	array := func(item any) map[string]any { return map[string]any{"type": "array", "items": item} }
	obj := func(properties map[string]any) map[string]any {
		required := []string{}
		for name := range properties {
			required = append(required, name)
		}

		sort.Strings(required)

		return map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
	}
	sources := array(str)
	// Keep enum size within provider schema limits for larger workspaces.
	// The use case still validates every returned source against the full registry.
	if len(sourceIDs) > 0 && len(sourceIDs) <= 250 {
		sources = array(map[string]any{"type": "string", "enum": sourceIDs})
	}

	return obj(map[string]any{
		"html":            map[string]any{"type": "string", "description": "Leitura complementar da análise em HTML autossuficiente e em português. A análise completa também deve estar nos campos estruturados: métricas, seções, gráficos com pontos numéricos, tabelas, recomendações e limitações. O backend monta a apresentação final e desenha os gráficos reais. Sem placeholders, JavaScript, scripts, iframes, formulários ou recursos externos."},
		"title":           map[string]any{"type": "string", "description": "Título descritivo que você deve criar para este relatório. Obrigatório e nunca vazio; não depende de existir um título nos dados."},
		"summary":         map[string]any{"type": "string", "description": "Resumo em português da análise solicitada. Obrigatório e nunca vazio. Se não houver dados, explique a ausência neste resumo."},
		"metrics":         array(obj(map[string]any{"label": str, "value": num, "unit": str})),
		"sections":        array(obj(map[string]any{"title": str, "body": str})),
		"charts":          array(obj(map[string]any{"title": str, "kind": map[string]any{"type": "string", "enum": []string{"bar", "line"}}, "points": array(obj(map[string]any{"label": str, "value": num}))})),
		"tables":          array(obj(map[string]any{"title": str, "columns": array(str), "rows": array(array(str))})),
		"recommendations": array(str), "missing_data": array(str), "source_ids": sources,
	})
}
