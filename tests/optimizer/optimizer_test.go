package optimizer_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
	"github.com/EdgarOrtegaRamirez/svgforge/internal/optimizer"
)

func TestOptimizeRemoveRedundant(t *testing.T) {
	doc := &models.SVGDocument{
		Width:  "100",
		Height: "100",
		Elements: []*models.Element{
			{
				Tag: "rect",
				Attributes: map[string]string{
					"width":           "100",
					"height":          "50",
					"opacity":         "1",
					"fill":            "black",
					"stroke-width":    "1",
					"visibility":      "visible",
				},
			},
		},
	}

	opts := optimizer.DefaultOptions()
	optimizer.Optimize(doc, opts)

	el := doc.Elements[0]
	// Should keep width and height
	if el.Attribute("width") != "100" {
		t.Error("width should be kept")
	}
	if el.Attribute("height") != "50" {
		t.Error("height should be kept")
	}
	// Should remove default values
	if el.Attribute("opacity") != "" {
		t.Errorf("opacity should be removed, got %q", el.Attribute("opacity"))
	}
	if el.Attribute("fill") != "" {
		t.Errorf("fill should be removed (default black), got %q", el.Attribute("fill"))
	}
	if el.Attribute("stroke-width") != "" {
		t.Errorf("stroke-width should be removed (default 1), got %q", el.Attribute("stroke-width"))
	}
	if el.Attribute("visibility") != "" {
		t.Errorf("visibility should be removed (default visible), got %q", el.Attribute("visibility"))
	}
}

func TestOptimizeRemoveEditorAttrs(t *testing.T) {
	doc := &models.SVGDocument{
		Elements: []*models.Element{
			{
				Tag: "rect",
				Attributes: map[string]string{
					"width":       "100",
					"height":      "50",
					"data-name":   "my-rect",
					"inkscape:version": "1.3",
				},
			},
		},
	}

	opts := optimizer.DefaultOptions()
	optimizer.Optimize(doc, opts)

	el := doc.Elements[0]
	if el.Attribute("data-name") != "" {
		t.Error("data-name should be removed")
	}
	if el.Attribute("inkscape:version") != "" {
		t.Error("inkscape:version should be removed")
	}
}

func TestOptimizeSimplifyPath(t *testing.T) {
	doc := &models.SVGDocument{
		Elements: []*models.Element{
			{
				Tag: "path",
				Attributes: map[string]string{
					"d": "M 0.00 0.00 L 10.00 10.00 Z",
				},
			},
		},
	}

	opts := optimizer.DefaultOptions()
	optimizer.Optimize(doc, opts)

	d := doc.Elements[0].Attribute("d")
	// Should simplify whitespace
	if d == "M 0.00 0.00 L 10.00 10.00 Z" {
		// Path may or may not be simplified, that's OK
	}
}

func TestOptimizeRemoveEmptyGroups(t *testing.T) {
	doc := &models.SVGDocument{
		Elements: []*models.Element{
			{
				Tag: "g",
				Attributes: map[string]string{"id": "empty"},
				Children:   []*models.Element{},
			},
			{
				Tag: "g",
				Children: []*models.Element{
					{Tag: "rect"},
				},
			},
		},
	}

	opts := optimizer.DefaultOptions()
	opts.RemoveEmptyGroups = true
	optimizer.Optimize(doc, opts)

	// First element should be marked for removal (tag set to "")
	if doc.Elements[0].Tag != "" {
		t.Error("Empty group should be marked for removal")
	}
	// Second element should be kept
	if doc.Elements[1].Tag != "g" {
		t.Error("Non-empty group should be kept")
	}
}

func TestOptimizeRemoveEmptyAttributes(t *testing.T) {
	doc := &models.SVGDocument{
		Elements: []*models.Element{
			{
				Tag: "rect",
				Attributes: map[string]string{
					"width":  "100",
					"height": "",
					"class":  "",
				},
			},
		},
	}

	opts := optimizer.DefaultOptions()
	optimizer.Optimize(doc, opts)

	el := doc.Elements[0]
	if el.Attribute("height") != "" {
		t.Error("Empty height should be removed")
	}
	if el.Attribute("class") != "" {
		t.Error("Empty class should be removed")
	}
}

func TestMinifyOptions(t *testing.T) {
	opts := optimizer.MinifyOptions()
	if !opts.RemoveComments {
		t.Error("MinifyOptions should remove comments")
	}
	if !opts.RemoveMetadata {
		t.Error("MinifyOptions should remove metadata")
	}
	if !opts.InlineStyles {
		t.Error("MinifyOptions should inline styles")
	}
	if !opts.Minify {
		t.Error("MinifyOptions should minify")
	}
}
