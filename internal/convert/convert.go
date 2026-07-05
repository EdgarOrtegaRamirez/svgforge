// Package convert provides SVG format conversion.
package convert

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
)

// ToDataURI converts an SVG document to a data: URI.
func ToDataURI(doc *models.SVGDocument) (string, error) {
	data, err := ToBytes(doc)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:image/svg+xml;base64,%s", encoded), nil
}

// ToDataURIEncoded converts SVG to a URL-encoded data URI.
func ToDataURIEncoded(doc *models.SVGDocument) (string, error) {
	data, err := ToBytes(doc)
	if err != nil {
		return "", err
	}
	encoded := escapeCSS(string(data))
	return fmt.Sprintf("data:image/svg+xml,%s", encoded), nil
}

// ToInlineHTML converts SVG to an inline HTML element.
func ToInlineHTML(doc *models.SVGDocument) (string, error) {
	data, err := ToBytes(doc)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToCSSSprite creates a CSS sprite sheet from multiple SVGs.
func ToCSSSprite(svgs map[string]*models.SVGDocument) (string, error) {
	var sb strings.Builder
	sb.WriteString("<svg xmlns=\"http://www.w3.org/2000/svg\" style=\"display:none\">\n")

	offset := 0
	symbols := make([]string, 0)

	for name, doc := range svgs {
		width := doc.Width
		height := doc.Height
		if width == "" {
			width = "24"
		}
		if height == "" {
			height = "24"
		}

		innerContent := extractInnerContent(doc)
		symbol := fmt.Sprintf("  <symbol id=\"%s\" viewBox=\"%s\" width=\"%s\" height=\"%s\">\n%s  </symbol>\n",
			sanitizeID(name), getViewBox(doc), width, height, innerContent)
		symbols = append(symbols, symbol)
		offset++
	}

	for _, s := range symbols {
		sb.WriteString(s)
	}

	sb.WriteString("</svg>")
	return sb.String(), nil
}

// ToMinified produces a minified SVG string.
func ToMinified(doc *models.SVGDocument) (string, error) {
	data, err := ToBytes(doc)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToFormatted produces a nicely formatted SVG string.
func ToFormatted(doc *models.SVGDocument) (string, error) {
	var sb strings.Builder
	sb.WriteString(xml.Header)
	sb.WriteString(renderFormatted(doc, 0))
	return sb.String(), nil
}

// ToBytes renders the SVG document to bytes.
func ToBytes(doc *models.SVGDocument) ([]byte, error) {
	var sb strings.Builder
	sb.WriteString(xml.Header)
	sb.WriteString(renderSVG(doc))
	return []byte(sb.String()), nil
}

func renderSVG(doc *models.SVGDocument) string {
	var sb strings.Builder
	sb.WriteString("<svg")
	if doc.Xmlns != "" {
		fmt.Fprintf(&sb, " xmlns=\"%s\"", doc.Xmlns)
	} else {
		sb.WriteString(" xmlns=\"http://www.w3.org/2000/svg\"")
	}
	if doc.Version != "" {
		fmt.Fprintf(&sb, " version=\"%s\"", doc.Version)
	}
	if doc.Width != "" {
		fmt.Fprintf(&sb, " width=\"%s\"", doc.Width)
	}
	if doc.Height != "" {
		fmt.Fprintf(&sb, " height=\"%s\"", doc.Height)
	}
	if doc.ViewBox != nil {
		fmt.Fprintf(&sb, " viewBox=\"%.4g %.4g %.4g %.4g\"",
			doc.ViewBox.MinX, doc.ViewBox.MinY, doc.ViewBox.Width, doc.ViewBox.Height)
	}
	sb.WriteString(">\n")

	if doc.Title != "" {
		fmt.Fprintf(&sb, "  <title>%s</title>\n", doc.Title)
	}
	if doc.Desc != "" {
		fmt.Fprintf(&sb, "  <desc>%s</desc>\n", doc.Desc)
	}

	// Render defs
	if len(doc.Defs) > 0 {
		sb.WriteString("  <defs>\n")
		for _, def := range doc.Defs {
			sb.WriteString("    ")
			sb.WriteString(renderElement(def))
			sb.WriteString("\n")
		}
		sb.WriteString("  </defs>\n")
	}

	// Render elements
	for _, el := range doc.Elements {
		sb.WriteString("  ")
		sb.WriteString(renderElement(el))
		sb.WriteString("\n")
	}

	sb.WriteString("</svg>\n")
	return sb.String()
}

func renderElement(el *models.Element) string {
	if el == nil || el.Tag == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<")
	sb.WriteString(el.Tag)

	// Write attributes
	for k, v := range el.Attributes {
		fmt.Fprintf(&sb, " %s=\"%s\"", k, escapeAttr(v))
	}

	if el.IsContainer() || len(el.Children) > 0 || el.Text != "" {
		sb.WriteString(">")
		if el.Text != "" {
			sb.WriteString(escapeText(el.Text))
		}
		for _, child := range el.Children {
			sb.WriteString(renderElement(child))
		}
		fmt.Fprintf(&sb, "</%s>", el.Tag)
	} else {
		sb.WriteString(" />")
	}

	return sb.String()
}

func renderFormatted(doc *models.SVGDocument, indent int) string {
	prefix := strings.Repeat("  ", indent)
	var sb strings.Builder
	sb.WriteString(prefix + "<svg")
	if doc.Xmlns != "" {
		fmt.Fprintf(&sb, " xmlns=\"%s\"", doc.Xmlns)
	} else {
		sb.WriteString(" xmlns=\"http://www.w3.org/2000/svg\"")
	}
	if doc.Width != "" {
		fmt.Fprintf(&sb, "\n%s  width=\"%s\"", prefix, doc.Width)
	}
	if doc.Height != "" {
		fmt.Fprintf(&sb, "\n%s  height=\"%s\"", prefix, doc.Height)
	}
	if doc.ViewBox != nil {
		fmt.Fprintf(&sb, "\n%s  viewBox=\"%.4g %.4g %.4g %.4g\"",
			prefix, doc.ViewBox.MinX, doc.ViewBox.MinY, doc.ViewBox.Width, doc.ViewBox.Height)
	}
	sb.WriteString(">\n")

	if doc.Title != "" {
		fmt.Fprintf(&sb, "%s  <title>%s</title>\n", prefix, doc.Title)
	}

	for _, el := range doc.Elements {
		sb.WriteString(renderFormattedElement(el, indent+1))
	}

	fmt.Fprintf(&sb, "%s</svg>\n", prefix)
	return sb.String()
}

func renderFormattedElement(el *models.Element, indent int) string {
	if el == nil || el.Tag == "" {
		return ""
	}

	prefix := strings.Repeat("  ", indent)
	var sb strings.Builder
	sb.WriteString(prefix + "<" + el.Tag)

	for k, v := range el.Attributes {
		fmt.Fprintf(&sb, " %s=\"%s\"", k, escapeAttr(v))
	}

	if el.IsContainer() || len(el.Children) > 0 || el.Text != "" {
		sb.WriteString(">")
		if el.Text != "" {
			sb.WriteString(escapeText(el.Text))
		}
		sb.WriteString("\n")
		for _, child := range el.Children {
			sb.WriteString(renderFormattedElement(child, indent+1))
		}
		fmt.Fprintf(&sb, "%s</%s>\n", prefix, el.Tag)
	} else {
		sb.WriteString(" />\n")
	}

	return sb.String()
}

func extractInnerContent(doc *models.SVGDocument) string {
	var sb strings.Builder
	for _, el := range doc.Elements {
		sb.WriteString("    ")
		sb.WriteString(renderElement(el))
		sb.WriteString("\n")
	}
	return sb.String()
}

func getViewBox(doc *models.SVGDocument) string {
	if doc.ViewBox != nil {
		return fmt.Sprintf("%.4g %.4g %.4g %.4g",
			doc.ViewBox.MinX, doc.ViewBox.MinY, doc.ViewBox.Width, doc.ViewBox.Height)
	}
	return "0 0 24 24"
}

func sanitizeID(s string) string {
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, "/", "-")
	return s
}

func escapeAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func escapeText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func escapeCSS(s string) string {
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "#", "%23")
	s = strings.ReplaceAll(s, "<", "%3C")
	s = strings.ReplaceAll(s, ">", "%3E")
	return s
}
