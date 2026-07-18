package stats_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
	"github.com/EdgarOrtegaRamirez/svgforge/internal/stats"
)

func TestAnalyzeEmptySVG(t *testing.T) {
	doc := &models.SVGDocument{
		Width:  "100",
		Height: "100",
	}
	s := stats.Analyze(doc)
	if s.TotalElements != 0 {
		t.Errorf("TotalElements = %d, want 0", s.TotalElements)
	}
	if s.Width != "100" {
		t.Errorf("Width = %q, want %q", s.Width, "100")
	}
}

func TestAnalyzeShapes(t *testing.T) {
	doc := &models.SVGDocument{
		Elements: []*models.Element{
			{Tag: "rect"},
			{Tag: "circle"},
			{Tag: "ellipse"},
			{Tag: "line"},
			{Tag: "path"},
		},
	}
	s := stats.Analyze(doc)
	if s.ShapeElements != 5 {
		t.Errorf("ShapeElements = %d, want 5", s.ShapeElements)
	}
}

func TestAnalyzeContainers(t *testing.T) {
	doc := &models.SVGDocument{
		Elements: []*models.Element{
			{Tag: "g"},
			{Tag: "defs"},
			{Tag: "symbol"},
		},
	}
	s := stats.Analyze(doc)
	if s.ContainerElements != 3 {
		t.Errorf("ContainerElements = %d, want 3", s.ContainerElements)
	}
}

func TestAnalyzeTagCounts(t *testing.T) {
	doc := &models.SVGDocument{
		Elements: []*models.Element{
			{Tag: "rect"},
			{Tag: "rect"},
			{Tag: "rect"},
			{Tag: "circle"},
		},
	}
	s := stats.Analyze(doc)
	if s.TagCounts["rect"] != 3 {
		t.Errorf("TagCounts[rect] = %d, want 3", s.TagCounts["rect"])
	}
	if s.TagCounts["circle"] != 1 {
		t.Errorf("TagCounts[circle] = %d, want 1", s.TagCounts["circle"])
	}
}

func TestAnalyzeClassCounts(t *testing.T) {
	doc := &models.SVGDocument{
		Elements: []*models.Element{
			{Tag: "rect", Attributes: map[string]string{"class": "a b"}},
			{Tag: "circle", Attributes: map[string]string{"class": "a"}},
		},
	}
	s := stats.Analyze(doc)
	if s.ClassCounts["a"] != 2 {
		t.Errorf("ClassCounts[a] = %d, want 2", s.ClassCounts["a"])
	}
	if s.ClassCounts["b"] != 1 {
		t.Errorf("ClassCounts[b] = %d, want 1", s.ClassCounts["b"])
	}
}

func TestAnalyzeFeatures(t *testing.T) {
	doc := &models.SVGDocument{
		Elements: []*models.Element{
			{Tag: "style"},
			{Tag: "animate"},
			{Tag: "filter"},
			{Tag: "linearGradient"},
			{Tag: "mask"},
			{Tag: "clipPath"},
		},
	}
	s := stats.Analyze(doc)
	if !s.HasStyles {
		t.Error("HasStyles should be true")
	}
	if !s.HasAnimations {
		t.Error("HasAnimations should be true")
	}
	if !s.HasFilters {
		t.Error("HasFilters should be true")
	}
	if !s.HasGradients {
		t.Error("HasGradients should be true")
	}
	if !s.HasMasks {
		t.Error("HasMasks should be true")
	}
	if !s.HasClipPaths {
		t.Error("HasClipPaths should be true")
	}
}

func TestAnalyzeMaxDepth(t *testing.T) {
	doc := &models.SVGDocument{
		Elements: []*models.Element{
			{
				Tag: "g",
				Children: []*models.Element{
					{
						Tag: "g",
						Children: []*models.Element{
							{Tag: "rect"},
						},
					},
				},
			},
		},
	}
	s := stats.Analyze(doc)
	if s.MaxDepth != 3 {
		t.Errorf("MaxDepth = %d, want 3", s.MaxDepth)
	}
}

func TestFormatText(t *testing.T) {
	s := &stats.Stats{
		Width:         "100",
		Height:        "100",
		TotalElements: 5,
		ShapeElements: 3,
		TagCounts:     map[string]int{"rect": 2, "circle": 1},
	}
	text := stats.FormatText(s)
	if text == "" {
		t.Error("FormatText should return non-empty string")
	}
}
