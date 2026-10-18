package reporthtml

import (
	"bytes"
	"fmt"
	"html"
	"strings"

	htmlparser "golang.org/x/net/html"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
)

const (
	ContentSecurityPolicy = "default-src 'none'; script-src 'none'; style-src 'unsafe-inline'; img-src data:; font-src 'none'; connect-src 'none'; object-src 'none'; frame-src 'none'; base-uri 'none'; form-action 'none'"
	maxHTMLBytes          = 1024 * 1024
	styleElement          = "style"
	limitationsSectionID  = "squadflow-report-limitations"
)

type Renderer struct{}

var _ report.DocumentRenderer = Renderer{}

// Render composes structured reports and sanitizes the self-contained HTML document.
func (Renderer) Render(content report.Content) (string, error) {
	source := content.HTML
	if strings.TrimSpace(source) == "" {
		source = editorialHTML(content, nil)
	}

	if len(source) > maxHTMLBytes {
		return "", fmt.Errorf("%w: documento HTML acima de 1 MiB", report.ErrInvalidReport)
	}

	root, err := htmlparser.Parse(strings.NewReader(source))
	if err != nil {
		return "", fmt.Errorf("%w: HTML inválido", report.ErrInvalidReport)
	}

	clean(root)

	if !containsID(root, documentID) && (len(content.Metrics)+len(content.Charts)+len(content.Sections)+len(content.Tables)+len(content.Recommendations) > 0) {
		root, err = htmlparser.Parse(strings.NewReader(editorialHTML(content, root)))
		if err != nil {
			return "", fmt.Errorf("%w: composição HTML inválida", report.ErrInvalidReport)
		}

		clean(root)
	}

	var output bytes.Buffer

	output.WriteString("<!doctype html><html lang=\"pt-BR\"><head><meta charset=\"utf-8\"><meta http-equiv=\"Content-Security-Policy\" content=\"")
	output.WriteString(html.EscapeString(ContentSecurityPolicy))
	output.WriteString("\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"><title>")
	output.WriteString(html.EscapeString(content.Title))
	output.WriteString("</title>")

	head := find(root, "head")
	body := find(root, "body")

	if head != nil {
		for n := head.FirstChild; n != nil; n = n.NextSibling {
			if n.Type == htmlparser.ElementNode && n.Data == styleElement && !hasID(n, "squadflow-report-theme") {
				if err = htmlparser.Render(&output, n); err != nil {
					return "", fmt.Errorf("render report HTML: %w", err)
				}
			}
		}
	}

	output.WriteString("<style id=\"squadflow-report-theme\">" + editorialCSS + "</style></head><body>")

	if body == nil || !hasContent(body) {
		return "", fmt.Errorf("%w: documento HTML sem conteúdo visível", report.ErrInvalidReport)
	}

	for n := body.FirstChild; n != nil; n = n.NextSibling {
		if err = htmlparser.Render(&output, n); err != nil {
			return "", fmt.Errorf("render report HTML: %w", err)
		}
	}

	if len(content.MissingData) > 0 {
		output.WriteString("<section id=\"" + limitationsSectionID + "\" aria-label=\"Limitações e dados ausentes\"><h2>Limitações e dados ausentes</h2><ul>")

		for _, limitation := range content.MissingData {
			output.WriteString("<li>" + html.EscapeString(limitation) + "</li>")
		}

		output.WriteString("</ul></section>")
	}

	output.WriteString("</body></html>")

	if output.Len() > maxHTMLBytes {
		return "", fmt.Errorf("%w: documento HTML acima de 1 MiB", report.ErrInvalidReport)
	}

	return output.String(), nil
}

func find(node *htmlparser.Node, name string) *htmlparser.Node {
	if node == nil {
		return nil
	}

	if node.Type == htmlparser.ElementNode && node.Data == name {
		return node
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result := find(child, name); result != nil {
			return result
		}
	}

	return nil
}

func hasContent(node *htmlparser.Node) bool {
	if node.Type == htmlparser.ElementNode && node.Data == styleElement {
		return false
	}

	if node.Type == htmlparser.TextNode && strings.TrimSpace(node.Data) != "" {
		return true
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if hasContent(child) {
			return true
		}
	}

	return false
}

func allowedElement(name string) bool {
	switch name {
	case "html", "head", "body", "title", styleElement, "main", "article", "section", "header", "footer", "aside", "nav", "h1", "h2", "h3", "h4", "h5", "h6", "p", "strong", "b", "em", "i", "small", "span", "div", "ul", "ol", "li", "dl", "dt", "dd", "table", "thead", "tbody", "tfoot", "tr", "td", "th", "colgroup", "col", "caption", "blockquote", "pre", "code", "br", "hr", "details", "summary", "a", "img", "figure", "figcaption", "svg", "g", "path", "rect", "circle", "ellipse", "line", "polyline", "polygon", "text", "tspan", "defs", "linearGradient", "radialGradient", "stop", "clipPath", "pattern":
		return true
	default:
		return false
	}
}

func allowedAttribute(attribute htmlparser.Attribute) bool {
	key := strings.ToLower(attribute.Key)

	if attribute.Namespace != "" {
		return false
	}

	if key == "href" {
		return strings.HasPrefix(attribute.Val, "#")
	}

	if key == "src" {
		return strings.HasPrefix(attribute.Val, "data:image/png;base64,") || strings.HasPrefix(attribute.Val, "data:image/jpeg;base64,") || strings.HasPrefix(attribute.Val, "data:image/webp;base64,")
	}

	if strings.HasPrefix(key, "aria-") {
		return true
	}

	switch key {
	case "id", "class", styleElement, "title", "lang", "dir", "role", "alt", "width", "height", "viewbox", "xmlns", "x", "y", "x1", "x2", "y1", "y2", "cx", "cy", "r", "rx", "ry", "d", "points", "fill", "fill-opacity", "fill-rule", "stroke", "stroke-width", "stroke-opacity", "stroke-dasharray", "stroke-linecap", "stroke-linejoin", "opacity", "transform", "text-anchor", "dominant-baseline", "font-size", "font-family", "font-weight", "offset", "stop-color", "stop-opacity", "gradientunits", "gradienttransform", "clip-path", "preserveaspectratio", "colspan", "rowspan", "scope":
		return true
	default:
		return false
	}
}

func clean(n *htmlparser.Node) {
	for child := n.FirstChild; child != nil; {
		next := child.NextSibling

		if child.Type == htmlparser.CommentNode || child.Type == htmlparser.DoctypeNode || (child.Type == htmlparser.ElementNode && (!allowedElement(child.Data) || generatedLimitations(child))) {
			n.RemoveChild(child)
		} else {
			attributes := child.Attr[:0]

			for _, attribute := range child.Attr {
				if allowedAttribute(attribute) {
					attributes = append(attributes, attribute)
				}
			}

			child.Attr = attributes
			clean(child)
		}

		child = next
	}
}

func generatedLimitations(node *htmlparser.Node) bool {
	for _, attribute := range node.Attr {
		if attribute.Key == "id" && attribute.Val == limitationsSectionID {
			return true
		}
	}

	return false
}

func hasID(node *htmlparser.Node, id string) bool {
	if node == nil {
		return false
	}

	for _, attribute := range node.Attr {
		if attribute.Key == "id" && attribute.Val == id {
			return true
		}
	}

	return false
}

func containsID(node *htmlparser.Node, id string) bool {
	if hasID(node, id) {
		return true
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if containsID(child, id) {
			return true
		}
	}

	return false
}
