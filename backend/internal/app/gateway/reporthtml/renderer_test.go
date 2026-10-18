package reporthtml

import (
	"strings"
	"testing"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
)

func TestHTMLPreservesLayoutAndBlocksActiveContent(t *testing.T) {
	t.Parallel()

	source := `<!doctype html><html><head><base href="https://bad.example/"><meta http-equiv="refresh" content="0;url=https://bad.example"><link rel="stylesheet" href="https://bad.example/x.css"><style>h1{color:purple}@media print{section{break-after:page}}</style></head><body onload="alert(1)"><h1>Entrega &amp; qualidade</h1><section><svg viewBox="0 0 100 100" onload="alert(2)"><rect width="80" height="20" fill="purple"/><text x="4" y="15">3 entregas</text><script>alert(3)</script><foreignObject><iframe src="https://bad.example"></iframe></foreignObject></svg></section><script>alert(4)</script><iframe srcdoc="x"></iframe><form action="https://bad.example"><input name="password"></form><a href="javascript:alert(5)" onclick="alert(6)">Evidências</a><img src="https://bad.example/tracker.png" onerror="alert(7)"></body></html>`

	document, err := (Renderer{}).Render(report.Content{Title: "Qualidade", Summary: "Resumo", HTML: source, MissingData: []string{"Coleta limitada a 2000 demandas"}})
	if err != nil {
		t.Fatal(err)
	}

	for _, forbidden := range []string{"<script", "<iframe", "<form", "<input", "<base", "<link", "onclick=", "onload=", "onerror=", "srcdoc=", "javascript:", "https://bad.example"} {
		if strings.Contains(document, forbidden) {
			t.Errorf("active content survived: %s", forbidden)
		}
	}

	for _, required := range []string{"Content-Security-Policy", "script-src", "default-src", "<svg", "<rect", "3 entregas", "@media print", "color:purple", "Entrega &amp; qualidade", "Coleta limitada a 2000 demandas"} {
		if !strings.Contains(document, required) {
			t.Errorf("report feature lost: %s", required)
		}
	}
}

func TestEmptyOrOversizedHTMLIsRejected(t *testing.T) {
	t.Parallel()

	for _, source := range []string{"<script>onlyUnsafe()</script>", strings.Repeat("a", maxHTMLBytes+1)} {
		if _, err := (Renderer{}).Render(report.Content{Title: "Relatório", HTML: source}); err == nil {
			t.Fatal("invalid document accepted")
		}
	}
}

func TestLegacyReportsCanBeExportedAsHTML(t *testing.T) {
	t.Parallel()

	document, err := (Renderer{}).Render(report.Content{Title: "Relatório antigo", Summary: "Resumo <script>alert(1)</script>", Metrics: []report.Metric{{Label: "Entregues", Value: 3, Unit: "demandas"}}, Tables: []report.Table{{Title: "Equipe", Columns: []string{"Nome"}, Rows: [][]string{{"Squad"}}}}, MissingData: []string{"Dados históricos indisponíveis"}})
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(document, "<script>") || !strings.Contains(document, "Squad") || !strings.Contains(document, "3 demandas") || !strings.Contains(document, "Dados históricos indisponíveis") {
		t.Fatalf("invalid legacy HTML: %s", document)
	}
}

func TestExportDoesNotDuplicateCollectedLimitations(t *testing.T) {
	t.Parallel()

	content := report.Content{Title: "Análise", Summary: "Resumo", HTML: "<main><h1>Análise</h1></main>", MissingData: []string{"Coleta limitada a 2000 demandas"}}

	document, err := (Renderer{}).Render(content)
	if err != nil {
		t.Fatal(err)
	}

	content.HTML = document

	exported, err := (Renderer{}).Render(content)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Count(exported, "Coleta limitada a 2000 demandas") != 1 {
		t.Fatal("export duplicated or omitted collected limitations")
	}
}

func TestStructuredChartsRenderRealValuesAndExportWithoutDuplication(t *testing.T) {
	t.Parallel()

	content := report.Content{
		Title: "Panorama das demandas", Summary: "Três demandas no recorte.",
		HTML:    "<h2>Análise complementar</h2><p>Contexto preservado.</p>",
		Metrics: []report.Metric{{Label: "Total", Value: 3, Unit: "demandas"}},
		Charts: []report.Chart{
			{Title: "Distribuição por status", Kind: "bar", Points: []report.Point{{Label: "Backlog", Value: 2}, {Label: "Em QA", Value: 1}, {Label: "Concluídas", Value: 0}}},
			{Title: "Evolução", Kind: "line", Points: []report.Point{{Label: "2026-10-16", Value: 0}, {Label: "2026-10-17", Value: 3}}},
		},
		MissingData: []string{"Sem demandas concluídas suficientes para estimar lead time."},
	}

	document, err := (Renderer{}).Render(content)
	if err != nil {
		t.Fatal(err)
	}

	for _, required := range []string{`width="370.00"`, `width="185.00"`, `width="0.00"`, `<polyline`, `<polygon`, `fill-opacity="0.18"`, `16/10/2026`, `17/10/2026`, `Contexto preservado.`, `sf-hero`, `Ver dados do gráfico`} {
		if !strings.Contains(document, required) {
			t.Errorf("missing rendered chart feature: %s", required)
		}
	}

	content.HTML = document

	exported, err := (Renderer{}).Render(content)
	if err != nil {
		t.Fatal(err)
	}

	for _, unique := range []string{`id="squadflow-document"`, `id="squadflow-report-theme"`, `Contexto preservado.`, content.MissingData[0]} {
		if strings.Count(exported, unique) != 1 {
			t.Errorf("export duplicated or omitted %s", unique)
		}
	}
}

func TestChartHandlesNegativeAndSingleValues(t *testing.T) {
	t.Parallel()

	for _, chart := range []report.Chart{
		{Title: "Variação", Kind: "bar", Points: []report.Point{{Label: "Redução <script>", Value: -2}, {Label: "Aumento", Value: 2}}},
		{Title: "Um dia", Kind: "line", Points: []report.Point{{Label: "2026-10-17", Value: 0}}},
	} {
		document, err := (Renderer{}).Render(report.Content{Title: "Variações", Summary: "Dados consultados", Charts: []report.Chart{chart}})
		if err != nil {
			t.Fatal(err)
		}

		for _, invalid := range []string{"NaN", "+Inf", "-Inf", "<script>"} {
			if strings.Contains(document, invalid) {
				t.Errorf("invalid chart output: %s", invalid)
			}
		}
	}
}
