package parser_test

import (
	"strings"
	"testing"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/parser"
)

func TestParseSimpleSVG(t *testing.T) {
	svg := `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
  <title>Test SVG</title>
  <rect x="10" y="10" width="80" height="80" fill="blue"/>
  <circle cx="50" cy="50" r="25" fill="red"/>
</svg>`

	p := parser.New()
	doc, err := p.Parse(strings.NewReader(svg))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if doc.Width != "100" {
		t.Errorf("Width = %q, want %q", doc.Width, "100")
	}
	if doc.Height != "100" {
		t.Errorf("Height = %q, want %q", doc.Height, "100")
	}
	if doc.Title != "Test SVG" {
		t.Errorf("Title = %q, want %q", doc.Title, "Test SVG")
	}
	if doc.ViewBox == nil {
		t.Fatal("ViewBox is nil")
	}
	if doc.ViewBox.Width != 100 {
		t.Errorf("ViewBox.Width = %v, want 100", doc.ViewBox.Width)
	}
	if len(doc.Elements) != 2 {
		t.Errorf("Elements count = %d, want 2", len(doc.Elements))
	}
}

func TestParseNestedElements(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg">
  <g id="group1">
    <rect width="100" height="50"/>
    <circle r="25"/>
    <g id="nested">
      <path d="M0,0 L10,10"/>
    </g>
  </g>
</svg>`

	p := parser.New()
	doc, err := p.Parse(strings.NewReader(svg))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(doc.Elements) != 1 {
		t.Fatalf("Elements count = %d, want 1", len(doc.Elements))
	}

	g := doc.Elements[0]
	if g.Tag != "g" {
		t.Errorf("Root element tag = %q, want %q", g.Tag, "g")
	}
	if g.ID() != "group1" {
		t.Errorf("Root group ID = %q, want %q", g.ID(), "group1")
	}
	if len(g.Children) != 3 {
		t.Errorf("Group children = %d, want 3", len(g.Children))
	}
}

func TestParseSelfClosingElements(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg">
  <rect x="0" y="0" width="100" height="100"/>
  <circle cx="50" cy="50" r="25"/>
  <line x1="0" y1="0" x2="100" y2="100"/>
  <path d="M0,0 L10,10 Z"/>
  <polyline points="0,0 10,10 20,0"/>
  <polygon points="0,0 10,10 20,0"/>
</svg>`

	p := parser.New()
	doc, err := p.Parse(strings.NewReader(svg))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(doc.Elements) != 6 {
		t.Errorf("Elements count = %d, want 6", len(doc.Elements))
	}

	expectedTags := []string{"rect", "circle", "line", "path", "polyline", "polygon"}
	for i, tag := range expectedTags {
		if doc.Elements[i].Tag != tag {
			t.Errorf("Element %d tag = %q, want %q", i, doc.Elements[i].Tag, tag)
		}
	}
}

func TestParseTextElements(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg">
  <text x="10" y="20">Hello</text>
  <g>
    <text x="30" y="40">World</text>
  </g>
</svg>`

	p := parser.New()
	doc, err := p.Parse(strings.NewReader(svg))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(doc.Elements) != 2 {
		t.Fatalf("Elements count = %d, want 2", len(doc.Elements))
	}

	if doc.Elements[0].Text != "Hello" {
		t.Errorf("First text = %q, want %q", doc.Elements[0].Text, "Hello")
	}

	g := doc.Elements[1]
	if g.Tag != "g" || len(g.Children) != 1 {
		t.Fatalf("Group should have 1 child")
	}
	if g.Children[0].Text != "World" {
		t.Errorf("Nested text = %q, want %q", g.Children[0].Text, "World")
	}
}

func TestParseAttributes(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg">
  <rect x="10" y="20" width="100" height="50" fill="blue" stroke="red" stroke-width="2"/>
</svg>`

	p := parser.New()
	doc, err := p.Parse(strings.NewReader(svg))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(doc.Elements) != 1 {
		t.Fatalf("Elements count = %d, want 1", len(doc.Elements))
	}

	rect := doc.Elements[0]
	if rect.Attribute("fill") != "blue" {
		t.Errorf("fill = %q, want %q", rect.Attribute("fill"), "blue")
	}
	if rect.Attribute("stroke") != "red" {
		t.Errorf("stroke = %q, want %q", rect.Attribute("stroke"), "red")
	}
	if rect.Attribute("stroke-width") != "2" {
		t.Errorf("stroke-width = %q, want %q", rect.Attribute("stroke-width"), "2")
	}
}

func TestParseEmptySVG(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg"></svg>`
	p := parser.New()
	doc, err := p.Parse(strings.NewReader(svg))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(doc.Elements) != 0 {
		t.Errorf("Elements count = %d, want 0", len(doc.Elements))
	}
}

func TestParseInvalidSVG(t *testing.T) {
	svg := `this is not svg`
	p := parser.New()
	_, err := p.Parse(strings.NewReader(svg))
	if err == nil {
		t.Error("Expected error for invalid SVG")
	}
}

func TestParseBytes(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect/></svg>`)
	p := parser.New()
	doc, err := p.ParseBytes(svg)
	if err != nil {
		t.Fatalf("ParseBytes error: %v", err)
	}
	if len(doc.Elements) != 1 {
		t.Errorf("Elements count = %d, want 1", len(doc.Elements))
	}
}
