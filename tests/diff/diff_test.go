package diff_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/diff"
	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
)

func TestDiffIdentical(t *testing.T) {
	doc1 := &models.SVGDocument{
		Width:  "100",
		Height: "100",
		Elements: []*models.Element{
			{Tag: "rect", Attributes: map[string]string{"width": "100"}},
		},
	}
	doc2 := &models.SVGDocument{
		Width:  "100",
		Height: "100",
		Elements: []*models.Element{
			{Tag: "rect", Attributes: map[string]string{"width": "100"}},
		},
	}

	result := diff.Diff(doc1, doc2)
	if result.Summary.Added != 0 || result.Summary.Removed != 0 || result.Summary.Modified != 0 {
		t.Errorf("Identical docs should have no changes: +%d -%d ~%d",
			result.Summary.Added, result.Summary.Removed, result.Summary.Modified)
	}
}

func TestDiffAddedElement(t *testing.T) {
	doc1 := &models.SVGDocument{
		Elements: []*models.Element{},
	}
	doc2 := &models.SVGDocument{
		Elements: []*models.Element{
			{Tag: "circle", Attributes: map[string]string{"r": "25"}},
		},
	}

	result := diff.Diff(doc1, doc2)
	if result.Summary.Added != 1 {
		t.Errorf("Added count = %d, want 1", result.Summary.Added)
	}
}

func TestDiffRemovedElement(t *testing.T) {
	doc1 := &models.SVGDocument{
		Elements: []*models.Element{
			{Tag: "rect"},
		},
	}
	doc2 := &models.SVGDocument{
		Elements: []*models.Element{},
	}

	result := diff.Diff(doc1, doc2)
	if result.Summary.Removed != 1 {
		t.Errorf("Removed count = %d, want 1", result.Summary.Removed)
	}
}

func TestDiffModifiedAttribute(t *testing.T) {
	doc1 := &models.SVGDocument{
		Elements: []*models.Element{
			{Tag: "rect", Attributes: map[string]string{"fill": "red"}},
		},
	}
	doc2 := &models.SVGDocument{
		Elements: []*models.Element{
			{Tag: "rect", Attributes: map[string]string{"fill": "blue"}},
		},
	}

	result := diff.Diff(doc1, doc2)
	if result.Summary.Modified == 0 {
		t.Error("Should detect modified attribute")
	}
}

func TestDiffDifferentWidth(t *testing.T) {
	doc1 := &models.SVGDocument{Width: "100"}
	doc2 := &models.SVGDocument{Width: "200"}

	result := diff.Diff(doc1, doc2)
	found := false
	for _, entry := range result.Entries {
		if entry.Path == "svg[@width]" && entry.Type == diff.ChangeModified {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should detect width change")
	}
}

func TestDiffFormatText(t *testing.T) {
	doc1 := &models.SVGDocument{Width: "100"}
	doc2 := &models.SVGDocument{Width: "200"}

	result := diff.Diff(doc1, doc2)
	text := diff.FormatText(result)
	if text == "" {
		t.Error("FormatText should return non-empty string")
	}
}

func TestDiffFormatCompact(t *testing.T) {
	doc1 := &models.SVGDocument{Width: "100"}
	doc2 := &models.SVGDocument{Width: "200"}

	result := diff.Diff(doc1, doc2)
	compact := diff.FormatCompact(result)
	if compact == "" {
		t.Error("FormatCompact should return non-empty string")
	}
}
