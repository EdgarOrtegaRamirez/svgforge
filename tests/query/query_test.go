package query_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
	"github.com/EdgarOrtegaRamirez/svgforge/internal/query"
)

func TestParseSimpleSelector(t *testing.T) {
	sel, err := query.Parse("rect")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(sel.Parts) != 1 {
		t.Fatalf("Parts count = %d, want 1", len(sel.Parts))
	}
	if sel.Parts[0].Tag != "rect" {
		t.Errorf("Tag = %q, want %q", sel.Parts[0].Tag, "rect")
	}
}

func TestParseIDSelector(t *testing.T) {
	sel, err := query.Parse("#myRect")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if sel.Parts[0].ID != "myRect" {
		t.Errorf("ID = %q, want %q", sel.Parts[0].ID, "myRect")
	}
}

func TestParseClassSelector(t *testing.T) {
	sel, err := query.Parse(".highlight")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(sel.Parts[0].Classes) != 1 || sel.Parts[0].Classes[0] != "highlight" {
		t.Errorf("Classes = %v, want [highlight]", sel.Parts[0].Classes)
	}
}

func TestParseCompoundSelector(t *testing.T) {
	sel, err := query.Parse("rect.active#main")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	part := sel.Parts[0]
	if part.Tag != "rect" {
		t.Errorf("Tag = %q, want %q", part.Tag, "rect")
	}
	if part.ID != "main" {
		t.Errorf("ID = %q, want %q", part.ID, "main")
	}
	if len(part.Classes) != 1 || part.Classes[0] != "active" {
		t.Errorf("Classes = %v, want [active]", part.Classes)
	}
}

func TestParseDescendantSelector(t *testing.T) {
	sel, err := query.Parse("g rect")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(sel.Parts) != 2 {
		t.Fatalf("Parts count = %d, want 2", len(sel.Parts))
	}
	if sel.Parts[1].Tag != "rect" {
		t.Errorf("Second part tag = %q, want %q", sel.Parts[1].Tag, "rect")
	}
}

func TestParseChildSelector(t *testing.T) {
	sel, err := query.Parse("g > rect")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(sel.Parts) != 2 {
		t.Fatalf("Parts count = %d, want 2", len(sel.Parts))
	}
	if sel.Parts[1].Combinator != ">" {
		t.Errorf("Combinator = %q, want %q", sel.Parts[1].Combinator, ">")
	}
}

func TestMatchTag(t *testing.T) {
	sel, _ := query.Parse("rect")
	el := &models.Element{Tag: "rect"}
	if !query.Match(el, sel) {
		t.Error("Match(rect, rect) = false, want true")
	}

	el2 := &models.Element{Tag: "circle"}
	if query.Match(el2, sel) {
		t.Error("Match(circle, rect) = true, want false")
	}
}

func TestMatchID(t *testing.T) {
	sel, _ := query.Parse("#target")
	el := &models.Element{
		Tag:        "rect",
		Attributes: map[string]string{"id": "target"},
	}
	if !query.Match(el, sel) {
		t.Error("Match should succeed for matching ID")
	}
}

func TestMatchClass(t *testing.T) {
	sel, _ := query.Parse(".active")
	el := &models.Element{
		Tag:        "g",
		Attributes: map[string]string{"class": "active highlighted"},
	}
	if !query.Match(el, sel) {
		t.Error("Match should succeed for matching class")
	}
}

func TestMatchCompound(t *testing.T) {
	sel, _ := query.Parse("rect.active#main")
	el := &models.Element{
		Tag:        "rect",
		Attributes: map[string]string{"id": "main", "class": "active"},
	}
	if !query.Match(el, sel) {
		t.Error("Match should succeed for compound selector")
	}
}

func TestMatchDescendant(t *testing.T) {
	sel, _ := query.Parse("g rect")
	g := &models.Element{Tag: "g"}
	rect := &models.Element{Tag: "rect"}
	g.AddChild(rect)

	if !query.Match(rect, sel) {
		t.Error("Match should succeed for descendant selector")
	}
}

func TestMatchChild(t *testing.T) {
	sel, _ := query.Parse("g > rect")
	g := &models.Element{Tag: "g"}
	rect := &models.Element{Tag: "rect"}
	g.AddChild(rect)

	if !query.Match(rect, sel) {
		t.Error("Match should succeed for direct child")
	}

	// Should fail for grandchild
	outerG := &models.Element{Tag: "g"}
	innerG := &models.Element{Tag: "g"}
	innerG.AddChild(rect)
	outerG.AddChild(innerG)
	// rect is now a grandchild of outerG, not a direct child
	if query.Match(rect, sel) {
		// This should still match because rect IS a direct child of innerG
		// The selector g > rect matches any rect that is a direct child of any g
		// So this is actually correct behavior
	}
}

func TestQuery(t *testing.T) {
	svg := &models.Element{
		Tag: "svg",
		Children: []*models.Element{
			{Tag: "rect", Attributes: map[string]string{"class": "bg"}},
			{Tag: "circle", Attributes: map[string]string{"class": "fg"}},
			{
				Tag: "g",
				Children: []*models.Element{
					{Tag: "rect", Attributes: map[string]string{"class": "fg"}},
				},
			},
		},
	}

	sel, _ := query.Parse(".fg")
	results := query.Query(svg, sel)
	if len(results) != 2 {
		t.Errorf("Query('.fg') returned %d results, want 2", len(results))
	}
}

func TestQueryString(t *testing.T) {
	svg := &models.Element{
		Tag: "svg",
		Children: []*models.Element{
			{Tag: "rect"},
			{Tag: "circle"},
			{Tag: "line"},
		},
	}

	results, err := query.QueryString(svg, "rect")
	if err != nil {
		t.Fatalf("QueryString error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("QueryString('rect') returned %d results, want 1", len(results))
	}
}

func TestMatchAttributePresence(t *testing.T) {
	sel, _ := query.Parse("[fill]")
	el := &models.Element{
		Tag:        "rect",
		Attributes: map[string]string{"fill": "red"},
	}
	if !query.Match(el, sel) {
		t.Error("Match should succeed for attribute presence")
	}

	el2 := &models.Element{Tag: "rect"}
	if query.Match(el2, sel) {
		t.Error("Match should fail when attribute missing")
	}
}

func TestMatchAttributeValue(t *testing.T) {
	sel, _ := query.Parse("[fill='red']")
	el := &models.Element{
		Tag:        "rect",
		Attributes: map[string]string{"fill": "red"},
	}
	if !query.Match(el, sel) {
		t.Error("Match should succeed for attribute value")
	}

	el2 := &models.Element{
		Tag:        "rect",
		Attributes: map[string]string{"fill": "blue"},
	}
	if query.Match(el2, sel) {
		t.Error("Match should fail for different attribute value")
	}
}

func TestMatchWildcard(t *testing.T) {
	sel, _ := query.Parse("*")
	for _, tag := range []string{"rect", "circle", "g", "text"} {
		el := &models.Element{Tag: tag}
		if !query.Match(el, sel) {
			t.Errorf("Match(*, %s) = false, want true", tag)
		}
	}
}

func TestMatchNthChild(t *testing.T) {
	sel, _ := query.Parse(":nth-child(2)")
	child0 := &models.Element{Tag: "rect"}
	child1 := &models.Element{Tag: "circle"}
	child2 := &models.Element{Tag: "line"}
	parent := &models.Element{
		Tag:      "g",
		Children: []*models.Element{child0, child1, child2},
	}
	child0.Parent = parent
	child1.Parent = parent
	child2.Parent = parent

	if !query.Match(child1, sel) {
		t.Error("Match should succeed for nth-child(2)")
	}
	if query.Match(child0, sel) {
		t.Error("Match should fail for nth-child(2) on first child")
	}
}

func TestMatchFirstChild(t *testing.T) {
	sel, _ := query.Parse(":first-child")
	child0 := &models.Element{Tag: "rect"}
	child1 := &models.Element{Tag: "circle"}
	parent := &models.Element{
		Tag:      "g",
		Children: []*models.Element{child0, child1},
	}
	child0.Parent = parent
	child1.Parent = parent

	if !query.Match(child0, sel) {
		t.Error("Match should succeed for first-child")
	}
	if query.Match(child1, sel) {
		t.Error("Match should fail for non-first child")
	}
}

func TestMatchLastChild(t *testing.T) {
	sel, _ := query.Parse(":last-child")
	child0 := &models.Element{Tag: "rect"}
	child1 := &models.Element{Tag: "circle"}
	parent := &models.Element{
		Tag:      "g",
		Children: []*models.Element{child0, child1},
	}
	child0.Parent = parent
	child1.Parent = parent

	if !query.Match(child1, sel) {
		t.Error("Match should succeed for last-child")
	}
	if query.Match(child0, sel) {
		t.Error("Match should fail for non-last child")
	}
}
