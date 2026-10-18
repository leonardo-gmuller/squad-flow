package reporthtml

import (
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
)

func chartHTML(chart report.Chart) string {
	var out strings.Builder

	out.WriteString("<figure class=\"sf-chart\"><figcaption>" + html.EscapeString(chart.Title) + "</figcaption>")

	if len(chart.Points) == 0 {
		out.WriteString("<p>Sem dados para este gráfico.</p></figure>")

		return out.String()
	}

	if chart.Kind == "line" {
		out.WriteString(lineSVG(chart))
	} else {
		out.WriteString(barSVG(chart))
	}

	out.WriteString("<details class=\"sf-chart-data\"><summary>Ver dados do gráfico</summary><table><thead><tr><th scope=\"col\">Categoria</th><th scope=\"col\">Valor</th></tr></thead><tbody>")

	for _, point := range chart.Points {
		out.WriteString("<tr><th scope=\"row\">" + html.EscapeString(chartLabel(point.Label)) + "</th><td>" + reportNumber(point.Value) + "</td></tr>")
	}

	out.WriteString("</tbody></table></details></figure>")

	return out.String()
}

func chartLabel(label string) string {
	if date, err := time.Parse("2006-01-02", label); err == nil {
		return date.Format("02/01/2006")
	}

	return label
}

func chartRange(points []report.Point) (float64, float64) {
	minimum, maximum := 0.0, 0.0

	for _, point := range points {
		minimum = math.Min(minimum, point.Value)
		maximum = math.Max(maximum, point.Value)
	}

	if minimum == maximum {
		maximum = minimum + 1
	}

	return minimum, maximum
}

func svgStart(chart report.Chart, height int) string {
	return fmt.Sprintf("<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 640 %d\" role=\"img\" aria-label=\"%s\" class=\"sf-chart-svg\"><title>%s</title>", height, html.EscapeString(chart.Title), html.EscapeString(chart.Title))
}

func barSVG(chart report.Chart) string {
	var out strings.Builder

	height := 48 + len(chart.Points)*62
	out.WriteString(svgStart(chart, height))

	minimum, maximum := chartRange(chart.Points)
	scale := func(value float64) float64 { return 200 + (value-minimum)/(maximum-minimum)*370 }
	zero := scale(0)

	for tick := range 5 {
		x := 200 + float64(tick)*92.5
		value := minimum + float64(tick)*(maximum-minimum)/4
		out.WriteString(fmt.Sprintf("<line x1=\"%.2f\" y1=\"12\" x2=\"%.2f\" y2=\"%d\" stroke=\"#e8e6f1\"/><text x=\"%.2f\" y=\"%d\" text-anchor=\"middle\" class=\"sf-axis\">%s</text>", x, x, height-30, x, height-8, reportNumber(value)))
	}

	colors := []string{"#7c3aed", "#0284c7", "#059669", "#d97706", "#e11d48"}

	for index, point := range chart.Points {
		rowY := 20 + index*62
		label := chartLabel(point.Label)
		display := []rune(label)

		if len(display) > 25 {
			label = string(display[:24]) + "…"
		}

		x := math.Min(zero, scale(point.Value))
		width := math.Abs(scale(point.Value) - zero)
		color := colors[index%len(colors)]
		out.WriteString(fmt.Sprintf("<text x=\"184\" y=\"%d\" text-anchor=\"end\" class=\"sf-category\"><title>%s</title>%s</text><rect x=\"%.2f\" y=\"%d\" width=\"%.2f\" height=\"30\" rx=\"6\" fill=\"%s\" fill-opacity=\"0.18\" stroke=\"%s\" stroke-width=\"2\"><title>%s: %s</title></rect><text x=\"610\" y=\"%d\" text-anchor=\"end\" class=\"sf-value\">%s</text>", rowY+21, html.EscapeString(point.Label), html.EscapeString(label), x, rowY, width, color, color, html.EscapeString(point.Label), reportNumber(point.Value), rowY+21, reportNumber(point.Value)))
	}

	out.WriteString("</svg>")

	return out.String()
}

func lineSVG(chart report.Chart) string {
	var out strings.Builder

	out.WriteString(svgStart(chart, 280))
	minimum, maximum := chartRange(chart.Points)
	yPosition := func(value float64) float64 { return 226 - (value-minimum)/(maximum-minimum)*190 }
	xPosition := func(index int) float64 {
		if len(chart.Points) == 1 {
			return 334
		}

		return 52 + float64(index)/float64(len(chart.Points)-1)*554
	}

	for tick := range 5 {
		value := minimum + float64(tick)*(maximum-minimum)/4
		out.WriteString(fmt.Sprintf("<line x1=\"52\" y1=\"%.2f\" x2=\"606\" y2=\"%.2f\" stroke=\"#e8e6f1\"/><text x=\"40\" y=\"%.2f\" text-anchor=\"end\" class=\"sf-axis\">%s</text>", yPosition(value), yPosition(value), yPosition(value)+4, reportNumber(value)))
	}

	var points strings.Builder

	for index, point := range chart.Points {
		fmt.Fprintf(&points, "%.2f,%.2f ", xPosition(index), yPosition(point.Value))
	}

	out.WriteString(fmt.Sprintf("<polygon points=\"%.2f,%.2f %s%.2f,%.2f\" fill=\"#7c3aed\" fill-opacity=\"0.12\"/><polyline points=\"%s\" fill=\"none\" stroke=\"#7c3aed\" stroke-width=\"3\" stroke-linejoin=\"round\"/>", xPosition(0), yPosition(0), points.String(), xPosition(len(chart.Points)-1), yPosition(0), points.String()))

	step := max(1, (len(chart.Points)+4)/5)

	for index, point := range chart.Points {
		out.WriteString(fmt.Sprintf("<circle cx=\"%.2f\" cy=\"%.2f\" r=\"4\" fill=\"#7c3aed\" stroke=\"white\" stroke-width=\"2\"><title>%s: %s</title></circle>", xPosition(index), yPosition(point.Value), html.EscapeString(chartLabel(point.Label)), reportNumber(point.Value)))

		if index%step == 0 || index == len(chart.Points)-1 {
			out.WriteString(fmt.Sprintf("<text x=\"%.2f\" y=\"258\" text-anchor=\"middle\" class=\"sf-axis\">%s</text>", xPosition(index), html.EscapeString(chartLabel(point.Label))))
		}
	}

	out.WriteString("</svg>")

	return out.String()
}
