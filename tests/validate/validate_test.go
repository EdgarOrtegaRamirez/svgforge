package validate_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
	"github.com/EdgarOrtegaRamirez/svgforge/internal/validate"
)

func TestValidateValidSVG(t *testing.T) {
	doc := &models.SVGDocument{
		Xmlns:  "http://www.w3.org/2000/svg",
		Width:  "100",
		Height: "100",
		Elements: []*models.Element{
			{Tag: "rect", Attributes: map[string]string{"width": "100", "height": "50"}},
		},
	}
	result := validate.Validate(doc)
	if result.Summary.Errors > 0 {
		t.Errorf("Valid SVG should have 0 errors, got %d", result.Summary.Errors)
	}
}

func TestValidateMissingXmlns(t *testing.T) {
	doc := &models.SVGDocument{
		Width:  "100",
		Height: "100",
	}
	result := validate.Validate(doc)
	found := false
	for _, issue := range result.Issues {
		if issue.Rule == "ROOT_NS" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should detect missing xmlns")
	}
}

func TestValidateEmptyPath(t *testing.T) {
	doc := &models.SVGDocument{
		Xmlns:  "http://www.w3.org/2000/svg",
		Width:  "100",
		Height: "100",
		Elements: []*models.Element{
			{Tag: "path", Attributes: map[string]string{"d": ""}},
		},
	}
	result := validate.Validate(doc)
	found := false
	for _, issue := range result.Issues {
		if issue.Rule == "PATH_EMPTY" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should detect empty path data")
	}
}

func TestValidateScriptElement(t *testing.T) {
	doc := &models.SVGDocument{
		Xmlns:  "http://www.w3.org/2000/svg",
		Width:  "100",
		Height: "100",
		Elements: []*models.Element{
			{Tag: "script", Text: "alert('xss')"},
		},
	}
	result := validate.Validate(doc)
	found := false
	for _, issue := range result.Issues {
		if issue.Rule == "SEC_SCRIPT" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should detect script element")
	}
}

func TestValidateEmptyGroup(t *testing.T) {
	doc := &models.SVGDocument{
		Xmlns:  "http://www.w3.org/2000/svg",
		Width:  "100",
		Height: "100",
		Elements: []*models.Element{
			{Tag: "g", Children: []*models.Element{}},
		},
	}
	result := validate.Validate(doc)
	found := false
	for _, issue := range result.Issues {
		if issue.Rule == "GROUP_EMPTY" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should detect empty group")
	}
}

func TestValidateFormatText(t *testing.T) {
	doc := &models.SVGDocument{
		Xmlns: "http://www.w3.org/2000/svg",
	}
	result := validate.Validate(doc)
	text := validate.FormatText(result)
	if text == "" {
		t.Error("FormatText should return non-empty string")
	}
}
