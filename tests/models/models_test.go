package models_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
)

func TestElement_Attribute(t *testing.T) {
	el := &models.Element{
		Tag:        "rect",
		Attributes: map[string]string{"width": "100", "height": "50"},
	}

	if got := el.Attribute("width"); got != "100" {
		t.Errorf("Attribute('width') = %q, want %q", got, "100")
	}
	if got := el.Attribute("nonexistent"); got != "" {
		t.Errorf("Attribute('nonexistent') = %q, want empty", got)
	}
}

func TestElement_SetAttribute(t *testing.T) {
	el := &models.Element{Tag: "circle"}
	el.SetAttribute("r", "25")
	if got := el.Attribute("r"); got != "25" {
		t.Errorf("SetAttribute failed: got %q, want %q", got, "25")
	}
}

func TestElement_HasClass(t *testing.T) {
	el := &models.Element{
		Tag:        "g",
		Attributes: map[string]string{"class": "foo bar baz"},
	}
	if !el.HasClass("foo") {
		t.Error("HasClass('foo') = false, want true")
	}
	if !el.HasClass("bar") {
		t.Error("HasClass('bar') = false, want true")
	}
	if el.HasClass("missing") {
		t.Error("HasClass('missing') = true, want false")
	}
}

func TestElement_AddClass(t *testing.T) {
	el := &models.Element{Tag: "g"}
	el.AddClass("new")
	if got := el.Attribute("class"); got != "new" {
		t.Errorf("AddClass on empty: got %q, want %q", got, "new")
	}
	el.AddClass("another")
	if got := el.Attribute("class"); got != "new another" {
		t.Errorf("AddClass with existing: got %q, want %q", got, "new another")
	}
	// Adding same class should not duplicate
	el.AddClass("new")
	if got := el.Attribute("class"); got != "new another" {
		t.Errorf("AddClass duplicate: got %q, want %q", got, "new another")
	}
}

func TestElement_RemoveClass(t *testing.T) {
	el := &models.Element{
		Tag:        "g",
		Attributes: map[string]string{"class": "foo bar baz"},
	}
	el.RemoveClass("bar")
	if got := el.Attribute("class"); got != "foo baz" {
		t.Errorf("RemoveClass: got %q, want %q", got, "foo baz")
	}
	// Remove last class should delete attribute
	el.RemoveClass("foo")
	el.RemoveClass("baz")
	if _, ok := el.Attributes["class"]; ok {
		t.Error("RemoveClass should delete attribute when empty")
	}
}

func TestElement_IsContainer(t *testing.T) {
	containerTests := []struct {
		tag  string
		want bool
	}{
		{"g", true},
		{"svg", true},
		{"defs", true},
		{"symbol", true},
		{"clipPath", true},
		{"mask", true},
		{"a", true},
		{"rect", false},
		{"circle", false},
		{"path", false},
		{"text", false},
	}

	for _, tt := range containerTests {
		el := &models.Element{Tag: tt.tag}
		if got := el.IsContainer(); got != tt.want {
			t.Errorf("IsContainer(%q) = %v, want %v", tt.tag, got, tt.want)
		}
	}
}

func TestElement_IsShape(t *testing.T) {
	shapeTests := []struct {
		tag  string
		want bool
	}{
		{"rect", true},
		{"circle", true},
		{"ellipse", true},
		{"line", true},
		{"polyline", true},
		{"polygon", true},
		{"path", true},
		{"text", true},
		{"g", false},
		{"defs", false},
	}

	for _, tt := range shapeTests {
		el := &models.Element{Tag: tt.tag}
		if got := el.IsShape(); got != tt.want {
			t.Errorf("IsShape(%q) = %v, want %v", tt.tag, got, tt.want)
		}
	}
}

func TestElement_FindByTag(t *testing.T) {
	root := &models.Element{
		Tag: "svg",
		Children: []*models.Element{
			{Tag: "rect", Attributes: map[string]string{"id": "r1"}},
			{
				Tag: "g",
				Children: []*models.Element{
					{Tag: "circle", Attributes: map[string]string{"id": "c1"}},
					{Tag: "rect", Attributes: map[string]string{"id": "r2"}},
				},
			},
		},
	}

	rects := root.FindByTag("rect")
	if len(rects) != 2 {
		t.Errorf("FindByTag('rect') returned %d results, want 2", len(rects))
	}

	circles := root.FindByTag("circle")
	if len(circles) != 1 {
		t.Errorf("FindByTag('circle') returned %d results, want 1", len(circles))
	}
}

func TestElement_FindByID(t *testing.T) {
	root := &models.Element{
		Tag: "svg",
		Children: []*models.Element{
			{Tag: "rect", Attributes: map[string]string{"id": "target"}},
			{Tag: "circle", Attributes: map[string]string{"id": "other"}},
		},
	}

	el := root.FindByID("target")
	if el == nil {
		t.Fatal("FindByID('target') returned nil")
	}
	if el.Tag != "rect" {
		t.Errorf("FindByID('target').Tag = %q, want %q", el.Tag, "rect")
	}

	if el := root.FindByID("nonexistent"); el != nil {
		t.Error("FindByID('nonexistent') should return nil")
	}
}

func TestElement_CountElements(t *testing.T) {
	root := &models.Element{
		Tag: "svg",
		Children: []*models.Element{
			{Tag: "rect"},
			{
				Tag: "g",
				Children: []*models.Element{
					{Tag: "circle"},
					{Tag: "line"},
				},
			},
		},
	}

	if got := root.CountElements(); got != 5 {
		t.Errorf("CountElements() = %d, want 5", got)
	}
}

func TestElement_Depth(t *testing.T) {
	root := &models.Element{
		Tag: "svg",
		Children: []*models.Element{
			{
				Tag: "g",
				Children: []*models.Element{
					{Tag: "rect"},
				},
			},
		},
	}

	if got := root.Depth(); got != 3 {
		t.Errorf("Depth() = %d, want 3", got)
	}
}

func TestElement_RemoveChild(t *testing.T) {
	child1 := &models.Element{Tag: "rect"}
	child2 := &models.Element{Tag: "circle"}
	parent := &models.Element{
		Tag:      "g",
		Children: []*models.Element{child1, child2},
	}

	if !parent.RemoveChild(child1) {
		t.Error("RemoveChild returned false, want true")
	}
	if len(parent.Children) != 1 {
		t.Errorf("After remove: %d children, want 1", len(parent.Children))
	}
	if parent.Children[0] != child2 {
		t.Error("Remaining child should be child2")
	}

	if parent.RemoveChild(child1) {
		t.Error("RemoveChild should return false for already removed child")
	}
}

func TestElement_AddChild(t *testing.T) {
	parent := &models.Element{Tag: "g"}
	child := &models.Element{Tag: "rect"}
	parent.AddChild(child)

	if len(parent.Children) != 1 {
		t.Errorf("AddChild: %d children, want 1", len(parent.Children))
	}
	if child.Parent != parent {
		t.Error("AddChild should set Parent")
	}
}

func TestElement_InsertBefore(t *testing.T) {
	child1 := &models.Element{Tag: "rect"}
	child2 := &models.Element{Tag: "circle"}
	ref := &models.Element{Tag: "line"}
	parent := &models.Element{
		Tag:      "g",
		Children: []*models.Element{child1, child2, ref},
	}

	newEl := &models.Element{Tag: "path"}
	if !parent.InsertBefore(ref, newEl) {
		t.Error("InsertBefore returned false")
	}
	if len(parent.Children) != 4 {
		t.Errorf("InsertBefore: %d children, want 4", len(parent.Children))
	}
	if parent.Children[2] != newEl {
		t.Error("New element should be at index 2")
	}
}

func TestViewBox(t *testing.T) {
	vb := &models.ViewBox{MinX: 0, MinY: 0, Width: 100, Height: 200}
	if vb.Width != 100 {
		t.Errorf("ViewBox.Width = %v, want 100", vb.Width)
	}
	if vb.Height != 200 {
		t.Errorf("ViewBox.Height = %v, want 200", vb.Height)
	}
}

func TestElement_String(t *testing.T) {
	el := &models.Element{
		Tag:        "rect",
		Attributes: map[string]string{"width": "100"},
	}
	s := el.String()
	if s != "<rect>" {
		t.Errorf("String() = %q, want %q", s, "<rect>")
	}

	textEl := &models.Element{
		Tag:  "text",
		Text: "Hello World",
	}
	s = textEl.String()
	if s != "<text>Hello World</text>" {
		t.Errorf("String() = %q, want %q", s, "<text>Hello World</text>")
	}

	longTextEl := &models.Element{
		Tag:  "text",
		Text: "This is a very long text that should be truncated at some point",
	}
	s = longTextEl.String()
	if len(s) > 50 {
		t.Errorf("String() for long text too long: %d chars", len(s))
	}
}

func TestElement_BBox(t *testing.T) {
	el := &models.Element{
		Tag: "rect",
		Attributes: map[string]string{
			"x":      "10",
			"y":      "20",
			"width":  "100",
			"height": "50",
		},
	}
	x, y, w, h := el.BBox()
	if x != 10 || y != 20 || w != 100 || h != 50 {
		t.Errorf("BBox() = (%v, %v, %v, %v), want (10, 20, 100, 50)", x, y, w, h)
	}
}

func TestElement_FindByClass(t *testing.T) {
	root := &models.Element{
		Tag: "svg",
		Children: []*models.Element{
			{Tag: "rect", Attributes: map[string]string{"class": "highlight"}},
			{Tag: "circle", Attributes: map[string]string{"class": "highlight"}},
			{Tag: "line", Attributes: map[string]string{"class": "normal"}},
		},
	}

	results := root.FindByClass("highlight")
	if len(results) != 2 {
		t.Errorf("FindByClass('highlight') returned %d results, want 2", len(results))
	}
}
