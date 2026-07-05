// Package parser provides SVG document parsing.
package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
)

// Parser parses SVG documents from XML.
type Parser struct {
	strict bool
}

// New creates a new SVG parser.
func New() *Parser {
	return &Parser{strict: false}
}

// Parse parses an SVG document from a reader.
func (p *Parser) Parse(r io.Reader) (*models.SVGDocument, error) {
	decoder := xml.NewDecoder(r)
	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose
	decoder.Entity = xml.HTMLEntity

	doc := &models.SVGDocument{
		Elements: make([]*models.Element, 0),
		Defs:     make([]*models.Element, 0),
		Metadata: make(map[string]string),
	}

	var stack []*models.Element
	var inSvg bool

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse error at line %d: %w", decoder.InputOffset(), err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			tag := normalizeTag(t.Name.Local)
			attrs := make(map[string]string)
			for _, attr := range t.Attr {
				attrs[attr.Name.Local] = attr.Value
			}

			line, col := decoder.InputPos() // approximate

			el := &models.Element{
				Tag:        tag,
				Attributes: attrs,
				Children:   make([]*models.Element, 0),
				Line:       line,
				Column:     col,
			}

			// Extract SVG-level attributes
			if !inSvg && tag == "svg" {
				inSvg = true
				doc.Xmlns = attrs["xmlns"]
				doc.Version = attrs["version"]
				doc.Width = attrs["width"]
				doc.Height = attrs["height"]
				if vb, ok := attrs["viewBox"]; ok {
					doc.ViewBox = parseViewBox(vb)
				}
			}

			// Handle self-closing elements
			if isSelfClosing(tag) {
				if len(stack) > 0 {
					stack[len(stack)-1].AddChild(el)
				} else if tag == "title" || tag == "desc" {
					// Self-closing title/desc elements — text content is handled by char data
					_ = el // element parsed but not attached to tree
				}
			} else {
				stack = append(stack, el)
			}

			// Handle defs
			if tag == "defs" && len(stack) > 0 {
				doc.Defs = append(doc.Defs, stack[len(stack)-1])
			}

		case xml.EndElement:
			tag := normalizeTag(t.Name.Local)
			if len(stack) == 0 {
				continue
			}

			current := stack[len(stack)-1]
			if normalizeTag(current.Tag) == tag {
				stack = stack[:len(stack)-1]
				if len(stack) > 0 {
					stack[len(stack)-1].AddChild(current)
				} else if normalizeTag(current.Tag) == "svg" {
					// Root svg element popped — collect its children
					// Exclude title and desc from elements
					for _, child := range current.Children {
						tag := normalizeTag(child.Tag)
						if tag != "title" && tag != "desc" {
							doc.Elements = append(doc.Elements, child)
						}
					}
				}
			} else {
				// Mismatched tags — try to find matching
				found := false
				for i := len(stack) - 1; i >= 0; i-- {
					if normalizeTag(stack[i].Tag) == tag {
						// Pop everything up to and including the match
						finished := stack[i]
						stack = stack[:i]
						if len(stack) > 0 {
							stack[len(stack)-1].AddChild(finished)
						}
						found = true
						break
					}
				}
				// In lenient mode, silently skip mismatched closing tags
				// In strict mode, this would be an error (handled elsewhere)
				_ = found // used for potential future strict-mode error handling
			}

		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text == "" || len(stack) == 0 {
				continue
			}
			current := stack[len(stack)-1]
			if normalizeTag(current.Tag) == "title" {
				doc.Title = text
			} else if normalizeTag(current.Tag) == "desc" {
				doc.Desc = text
			} else {
				current.Text = text
			}

		case xml.Comment:
			// Skip comments
		}
	}

	// Collect top-level elements (for malformed SVGs without closing tag)
	if len(stack) > 0 {
		for _, el := range stack {
			if normalizeTag(el.Tag) != "svg" {
				doc.Elements = append(doc.Elements, el)
			}
		}
	}

	// Validate that we found an SVG document
	if !inSvg && len(doc.Elements) == 0 {
		return nil, fmt.Errorf("not a valid SVG document: no <svg> root element found")
	}

	return doc, nil
}

// ParseBytes parses SVG from a byte slice.
func (p *Parser) ParseBytes(data []byte) (*models.SVGDocument, error) {
	return p.Parse(strings.NewReader(string(data)))
}

// ParseFile parses SVG from a file path.
func (p *Parser) ParseFile(path string) (*models.SVGDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return p.ParseBytes(data)
}

// parseViewBox parses a viewBox string like "0 0 100 200".
func parseViewBox(s string) *models.ViewBox {
	vb := &models.ViewBox{}
	_, _ = fmt.Sscanf(s, "%f %f %f %f", &vb.MinX, &vb.MinY, &vb.Width, &vb.Height)
	return vb
}

// normalizeTag normalizes SVG tag names.
func normalizeTag(tag string) string {
	// Handle namespace prefixes
	if idx := strings.Index(tag, ":"); idx >= 0 {
		tag = tag[idx+1:]
	}
	return strings.ToLower(tag)
}

// isSelfClosing checks if an SVG element is typically self-closing.
func isSelfClosing(tag string) bool {
	switch tag {
	case "rect", "circle", "ellipse", "line", "polyline", "polygon",
		"image", "use", "br", "hr", "input", "meta", "link",
		"stop", "animate", "animateTransform", "animateMotion",
		"feBlend", "feColorMatrix", "feComponentTransfer", "feComposite",
		"feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap",
		"feDistantLight", "feDropShadow", "feFlood", "feGaussianBlur",
		"feImage", "feMerge", "feMergeNode", "feMorphology", "feOffset",
		"fePointLight", "feSpecularLighting", "feSpotLight", "feTile", "feTurbulence",
		"set", "metadata":
		return true
	}
	return false
}
