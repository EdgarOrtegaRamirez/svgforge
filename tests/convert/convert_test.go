package convert_test

import (
	"strings"
	"testing"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/convert"
	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
)

func TestToBytes(t *testing.T) {
	doc := &models.SVGDocument{
		Xmlns:  "http://www.w3.org/2000/svg",
		Width:  "100",
		Height: "100",
		Elements: []*models.Element{
			{Tag: "rect", Attributes: map[string]string{"width": "100"}},
		},
	}

	data, err := convert.ToBytes(doc)
	if err != nil {
		t.Fatalf("ToBytes error: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, "<svg") {
		t.Error("Output should contain <svg")
	}
	if !strings.Contains(s, "xmlns") {
		t.Error("Output should contain xmlns")
	}
	if !strings.Contains(s, "<rect") {
		t.Error("Output should contain <rect")
	}
}

func TestToDataURI(t *testing.T) {
	doc := &models.SVGDocument{
		Xmlns: "http://www.w3.org/2000/svg",
	}

	uri, err := convert.ToDataURI(doc)
	if err != nil {
		t.Fatalf("ToDataURI error: %v", err)
	}

	if !strings.HasPrefix(uri, "data:image/svg+xml;base64,") {
		t.Errorf("URI should start with data:image/svg+xml;base64,, got %q", uri[:30])
	}
}

func TestToDataURIEncoded(t *testing.T) {
	doc := &models.SVGDocument{
		Xmlns: "http://www.w3.org/2000/svg",
	}

	uri, err := convert.ToDataURIEncoded(doc)
	if err != nil {
		t.Fatalf("ToDataURIEncoded error: %v", err)
	}

	if !strings.HasPrefix(uri, "data:image/svg+xml,") {
		t.Errorf("URI should start with data:image/svg+xml,, got %q", uri[:20])
	}
}

func TestToInlineHTML(t *testing.T) {
	doc := &models.SVGDocument{
		Xmlns: "http://www.w3.org/2000/svg",
	}

	html, err := convert.ToInlineHTML(doc)
	if err != nil {
		t.Fatalf("ToInlineHTML error: %v", err)
	}

	if !strings.Contains(html, "<svg") {
		t.Error("HTML should contain <svg")
	}
}

func TestToFormatted(t *testing.T) {
	doc := &models.SVGDocument{
		Xmlns:  "http://www.w3.org/2000/svg",
		Width:  "100",
		Height: "100",
		Elements: []*models.Element{
			{Tag: "rect", Attributes: map[string]string{"width": "100"}},
		},
	}

	s, err := convert.ToFormatted(doc)
	if err != nil {
		t.Fatalf("ToFormatted error: %v", err)
	}

	if !strings.Contains(s, "<?xml") {
		t.Error("Formatted should contain XML declaration")
	}
	if !strings.Contains(s, "  <rect") {
		t.Error("Formatted should have indentation")
	}
}

func TestToMinified(t *testing.T) {
	doc := &models.SVGDocument{
		Xmlns: "http://www.w3.org/2000/svg",
	}

	s, err := convert.ToMinified(doc)
	if err != nil {
		t.Fatalf("ToMinified error: %v", err)
	}

	if !strings.Contains(s, "<svg") {
		t.Error("Minified should contain <svg")
	}
}

func TestToCSSSprite(t *testing.T) {
	svgs := map[string]*models.SVGDocument{
		"icon-home": {
			Xmlns:  "http://www.w3.org/2000/svg",
			Width:  "24",
			Height: "24",
		},
		"icon-user": {
			Xmlns:  "http://www.w3.org/2000/svg",
			Width:  "32",
			Height: "32",
		},
	}

	sprite, err := convert.ToCSSSprite(svgs)
	if err != nil {
		t.Fatalf("ToCSSSprite error: %v", err)
	}

	if !strings.Contains(sprite, "<svg") {
		t.Error("Sprite should contain <svg")
	}
	if !strings.Contains(sprite, "<symbol") {
		t.Error("Sprite should contain <symbol>")
	}
	if !strings.Contains(sprite, "icon-home") {
		t.Error("Sprite should contain icon-home")
	}
	if !strings.Contains(sprite, "icon-user") {
		t.Error("Sprite should contain icon-user")
	}
}
