// Package validate provides SVG validation.
package validate

import (
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
)

// Severity represents the severity of a validation issue.
type Severity int

const (
	Info Severity = iota
	Warning
	Error
)

// Issue represents a validation issue.
type Issue struct {
	Severity Severity
	Path     string
	Message  string
	Rule     string
}

// ValidationResult holds all validation issues.
type ValidationResult struct {
	Issues  []Issue
	Summary ValidationSummary
}

// ValidationSummary provides counts of issues by severity.
type ValidationSummary struct {
	Errors   int
	Warnings int
	Infos    int
}

// Validate performs comprehensive SVG validation.
func Validate(doc *models.SVGDocument) *ValidationResult {
	result := &ValidationResult{}

	// Check root element
	validateRoot(doc, result)

	// Check for required elements
	validateRequired(doc, result)

	// Validate all elements
	validateElements(doc.Elements, "", result)

	// Check accessibility
	validateAccessibility(doc, result)

	// Check for common issues
	validateCommonIssues(doc, result)

	// Calculate summary
	for _, issue := range result.Issues {
		switch issue.Severity {
		case Error:
			result.Summary.Errors++
		case Warning:
			result.Summary.Warnings++
		case Info:
			result.Summary.Infos++
		}
	}

	return result
}

func validateRoot(doc *models.SVGDocument, result *ValidationResult) {
	if doc.Xmlns == "" {
		addIssue(result, Error, "svg", "Missing xmlns attribute", "ROOT_NS")
	}
	if doc.Width == "" && doc.ViewBox == nil {
		addIssue(result, Warning, "svg", "No width or viewBox specified", "ROOT_SIZE")
	}
	if doc.Height == "" && doc.ViewBox == nil {
		addIssue(result, Warning, "svg", "No height or viewBox specified", "ROOT_SIZE")
	}
}

func validateRequired(doc *models.SVGDocument, result *ValidationResult) {
	hasTitle := false
	for _, el := range doc.Elements {
		if el.Tag == "title" {
			hasTitle = true
			break
		}
	}
	if !hasTitle {
		addIssue(result, Info, "svg", "No <title> element found (recommended for accessibility)", "A11Y_TITLE")
	}
}

func validateElements(elements []*models.Element, basePath string, result *ValidationResult) {
	for _, el := range elements {
		if el == nil || el.Tag == "" {
			continue
		}
		path := basePath + "/" + el.Tag
		if id := el.ID(); id != "" {
			path += "[@" + id + "]"
		}

		validateElement(el, path, result)
		validateElements(el.Children, path, result)
	}
}

func validateElement(el *models.Element, path string, result *ValidationResult) {
	// Check for deprecated attributes
	deprecatedAttrs := []string{"xlink:href"}
	for _, attr := range deprecatedAttrs {
		if _, ok := el.Attributes[attr]; ok {
			addIssue(result, Warning, path, fmt.Sprintf("Deprecated attribute: %s", attr), "DEPRECATED")
		}
	}

	// Validate path data
	if el.Tag == "path" {
		d := el.Attribute("d")
		if d == "" {
			addIssue(result, Warning, path, "Path element has empty d attribute", "PATH_EMPTY")
		} else if !isValidPathData(d) {
			addIssue(result, Error, path, "Invalid path data", "PATH_INVALID")
		}
	}

	// Validate circle
	if el.Tag == "circle" {
		r := el.Attribute("r")
		if r == "" {
			addIssue(result, Warning, path, "Circle missing radius", "CIRCLE_R")
		}
	}

	// Validate ellipse
	if el.Tag == "ellipse" {
		rx := el.Attribute("rx")
		ry := el.Attribute("ry")
		if rx == "" || ry == "" {
			addIssue(result, Warning, path, "Ellipse missing rx or ry", "ELLIPSE_R")
		}
	}

	// Validate text
	if el.Tag == "text" {
		if el.Text == "" && len(el.Children) == 0 {
			addIssue(result, Info, path, "Text element has no content", "TEXT_EMPTY")
		}
	}

	// Validate viewBox references
	if el.Tag == "use" {
		href := el.Attribute("href")
		if href == "" {
			href = el.Attribute("xlink:href")
		}
		if href == "" {
			addIssue(result, Warning, path, "Use element missing href", "USE_HREF")
		}
	}

	// Check for empty groups
	if el.Tag == "g" && len(el.Children) == 0 {
		addIssue(result, Info, path, "Empty group element", "GROUP_EMPTY")
	}

	// Check for inline styles (should use classes)
	if style := el.Attribute("style"); style != "" {
		addIssue(result, Info, path, "Inline style detected — consider using CSS classes", "INLINE_STYLE")
	}

	// Check for very large coordinates
	for _, attr := range []string{"x", "y", "cx", "cy", "r", "rx", "ry"} {
		if val := el.Attribute(attr); val != "" {
			// Just check for obviously large numbers
			if len(val) > 10 {
				addIssue(result, Info, path, fmt.Sprintf("Large coordinate value in %s: %s", attr, val), "LARGE_COORD")
			}
		}
	}
}

func validateAccessibility(doc *models.SVGDocument, result *ValidationResult) {
	// Check for role attribute
	hasRole := false
	for _, el := range doc.Elements {
		if el.Attribute("role") != "" {
			hasRole = true
			break
		}
	}
	if !hasRole {
		addIssue(result, Info, "svg", "No role attribute found (recommended for accessibility)", "A11Y_ROLE")
	}

	// Check for aria attributes
	hasAria := false
	for _, el := range doc.Elements {
		for attr := range el.Attributes {
			if strings.HasPrefix(attr, "aria-") {
				hasAria = true
				break
			}
		}
		if hasAria {
			break
		}
	}
	if !hasAria {
		addIssue(result, Info, "svg", "No ARIA attributes found (recommended for accessibility)", "A11Y_ARIA")
	}
}

func validateCommonIssues(doc *models.SVGDocument, result *ValidationResult) {
	// Check for scripts
	for _, el := range doc.Elements {
		if el.Tag == "script" {
			addIssue(result, Warning, "svg", "Script element detected — potential security risk", "SEC_SCRIPT")
		}
	}

	// Check for foreign elements
	for _, el := range doc.Elements {
		if el.Tag == "foreignObject" {
			addIssue(result, Warning, "svg", "ForeignObject element — may not render in all viewers", "FOREIGN")
		}
	}
}

func isValidPathData(d string) bool {
	// Basic validation: should not be empty, should start with a command letter
	d = strings.TrimSpace(d)
	if d == "" {
		return false
	}
	// Valid SVG path commands start with M, L, C, Q, S, T, A, Z (case insensitive)
	first := strings.ToUpper(d[:1])
	validCommands := "MLCQSTAZ"
	return strings.Contains(validCommands, first)
}

func addIssue(result *ValidationResult, severity Severity, path, message, rule string) {
	result.Issues = append(result.Issues, Issue{
		Severity: severity,
		Path:     path,
		Message:  message,
		Rule:     rule,
	})
}

// FormatText formats validation results as text.
func FormatText(result *ValidationResult) string {
	var sb strings.Builder
	sb.WriteString("SVG Validation Report\n")
	sb.WriteString("=====================\n\n")

	for _, issue := range result.Issues {
		var sev string
		switch issue.Severity {
		case Error:
			sev = "ERROR"
		case Warning:
			sev = "WARN "
		case Info:
			sev = "INFO "
		}
		fmt.Fprintf(&sb, "[%s] %s: %s (%s)\n", sev, issue.Path, issue.Message, issue.Rule)
	}

	fmt.Fprintf(&sb, "\nSummary: %d errors, %d warnings, %d info\n",
		result.Summary.Errors, result.Summary.Warnings, result.Summary.Infos)

	if result.Summary.Errors == 0 {
		sb.WriteString("\n✓ SVG is valid\n")
	} else {
		sb.WriteString("\n✗ SVG has errors\n")
	}

	return sb.String()
}
