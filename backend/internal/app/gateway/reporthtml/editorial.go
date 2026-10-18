package reporthtml

import (
	"bytes"
	"fmt"
	"html"
	"strings"

	htmlparser "golang.org/x/net/html"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"

	_ "embed"
)

//go:embed report.css
var editorialCSS string

const documentID = "squadflow-document"

// The structured response is the source of truth for the document's visual components.
// Model-authored HTML remains available as a complementary reading, never as a chart engine.
func editorialHTML(content report.Content, original *htmlparser.Node) string {
	var out strings.Builder

	out.WriteString("<html><head></head><body><main id=\"" + documentID + "\" class=\"sf-document\"><header class=\"sf-hero\"><div class=\"sf-brand\"><span class=\"sf-logo\">S</span> SquadFlow <span class=\"sf-edition\">RELATÓRIO EXECUTIVO</span></div><div class=\"sf-kicker\">INTELIGÊNCIA DO SEU WORKSPACE</div><h1>" + html.EscapeString(content.Title) + "</h1><p>" + html.EscapeString(content.Summary) + "</p><div class=\"sf-hero-foot\"><span>Análise de desenvolvimento de software</span><span>Dados · Contexto · Decisões</span></div></header>")

	if len(content.Metrics) > 0 {
		out.WriteString("<section class=\"sf-metrics\" aria-label=\"Indicadores\">")

		for _, metric := range content.Metrics {
			out.WriteString("<article class=\"sf-metric\" aria-label=\"" + html.EscapeString(metric.Label+": "+reportNumber(metric.Value)+" "+metric.Unit) + "\"><span>" + html.EscapeString(metric.Label) + "</span><strong>" + reportNumber(metric.Value) + "<small> " + html.EscapeString(metric.Unit) + "</small></strong></article>")
		}

		out.WriteString("</section>")
	}

	if len(content.Charts) > 0 {
		out.WriteString("<section class=\"sf-section\"><div class=\"sf-section-heading\"><span class=\"sf-kicker\">LEITURA VISUAL</span><h2>Os dados em perspectiva</h2></div><div class=\"sf-charts\">")

		for _, chart := range content.Charts {
			out.WriteString(chartHTML(chart))
		}

		out.WriteString("</div></section>")
	}

	if len(content.Sections) > 0 {
		out.WriteString("<div class=\"sf-analysis\">")

		for index, section := range content.Sections {
			out.WriteString("<section class=\"sf-card\"><span class=\"sf-section-number\">" + fmt.Sprintf("%02d", index+1) + "</span><h2>" + html.EscapeString(section.Title) + "</h2><p class=\"sf-prose\">" + html.EscapeString(section.Body) + "</p></section>")
		}

		out.WriteString("</div>")
	}

	for _, table := range content.Tables {
		out.WriteString("<section class=\"sf-card sf-table-card\"><span class=\"sf-kicker\">DETALHAMENTO</span><h2>" + html.EscapeString(table.Title) + "</h2><div class=\"sf-table-scroll\"><table><caption class=\"sf-sr-only\">" + html.EscapeString(table.Title) + "</caption><thead><tr>")

		for _, column := range table.Columns {
			out.WriteString("<th scope=\"col\">" + html.EscapeString(column) + "</th>")
		}

		out.WriteString("</tr></thead><tbody>")

		for _, row := range table.Rows {
			out.WriteString("<tr>")

			for _, cell := range row {
				out.WriteString("<td>" + html.EscapeString(cell) + "</td>")
			}

			out.WriteString("</tr>")
		}

		out.WriteString("</tbody></table></div></section>")
	}

	if len(content.Recommendations) > 0 {
		out.WriteString("<section class=\"sf-card sf-actions\"><span class=\"sf-kicker\">DA ANÁLISE À AÇÃO</span><h2>Próximos passos</h2><ol>")

		for _, item := range content.Recommendations {
			out.WriteString("<li>" + html.EscapeString(item) + "</li>")
		}

		out.WriteString("</ol></section>")
	}

	if body := find(original, "body"); body != nil && hasContent(body) && strings.TrimSpace(content.HTML) != "" {
		out.WriteString("<details class=\"sf-card sf-complement\"><summary>Leitura complementar da análise</summary><div class=\"sf-complement-content\">")

		var supplemental bytes.Buffer

		for child := body.FirstChild; child != nil; child = child.NextSibling {
			_ = htmlparser.Render(&supplemental, child)
		}

		out.WriteString(supplemental.String())
		out.WriteString("</div></details>")
	}

	out.WriteString("<footer class=\"sf-footer\"><strong>SquadFlow</strong><span>Análise gerada por IA. Confira as evidências antes de tomar decisões.</span></footer></main></body></html>")

	return out.String()
}

func reportNumber(value float64) string {
	return strings.ReplaceAll(strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), "."), ".", ",")
}
