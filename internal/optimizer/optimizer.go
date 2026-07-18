// Package optimizer provides SVG optimization.
package optimizer

import (
	"regexp"
	"strings"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
)

// Options controls optimization behavior.
type Options struct {
	RemoveComments    bool
	RemoveMetadata    bool
	RemoveEditors     bool
	RemoveEmptyDefs   bool
	RemoveEmptyGroups bool
	SimplifyPaths     bool
	RemoveRedundant   bool
	InlineStyles      bool
	RemoveWhitespace  bool
	Minify            bool
}

// DefaultOptions returns sensible default optimization options.
func DefaultOptions() Options {
	return Options{
		RemoveComments:    true,
		RemoveMetadata:    true,
		RemoveEditors:     true,
		RemoveEmptyDefs:   true,
		RemoveEmptyGroups: true,
		SimplifyPaths:     true,
		RemoveRedundant:   true,
		InlineStyles:      false,
		RemoveWhitespace:  true,
		Minify:            false,
	}
}

// MinifyOptions returns aggressive optimization options.
func MinifyOptions() Options {
	opts := DefaultOptions()
	opts.Minify = true
	opts.InlineStyles = true
	return opts
}

// Optimize optimizes an SVG document.
func Optimize(doc *models.SVGDocument, opts Options) {
	// Remove metadata
	if opts.RemoveMetadata {
		doc.Metadata = make(map[string]string)
	}

	// Optimize all elements
	optimizeElements(doc.Elements, opts)

	// Clean up empty defs
	if opts.RemoveEmptyDefs {
		cleanDefs(doc)
	}
}

func optimizeElements(elements []*models.Element, opts Options) {
	for _, el := range elements {
		optimizeElement(el, opts)
		optimizeElements(el.Children, opts)
	}
}

func optimizeElement(el *models.Element, opts Options) {
	if el.Attributes == nil {
		return
	}

	if opts.RemoveRedundant {
		removeRedundantAttributes(el)
	}

	if opts.RemoveWhitespace {
		removeWhitespaceAttributes(el)
	}

	if opts.RemoveEditors {
		removeEditorAttributes(el)
	}

	if opts.SimplifyPaths && el.Tag == "path" {
		simplifyPath(el)
	}

	// Remove empty groups
	if opts.RemoveEmptyGroups && el.Tag == "g" && len(el.Children) == 0 && el.Text == "" {
		el.Tag = "" // mark for removal
	}
}

// removeRedundantAttributes removes default/empty attributes.
func removeRedundantAttributes(el *models.Element) {
	// Remove empty attributes
	for key, val := range el.Attributes {
		if val == "" {
			delete(el.Attributes, key)
		}
	}

	// Remove default values
	defaults := map[string]string{
		"opacity":          "1",
		"fill-opacity":     "1",
		"stroke-opacity":   "1",
		"stroke-width":     "1",
		"stroke-dasharray": "none",
		"stroke-linecap":   "butt",
		"stroke-linejoin":  "miter",
		"fill":             "black",
		"visibility":       "visible",
		"display":          "inline",
		"overflow":         "hidden",
	}

	for attr, defaultVal := range defaults {
		if val, ok := el.Attributes[attr]; ok && val == defaultVal {
			delete(el.Attributes, attr)
		}
	}

	// Remove xmlns on non-root elements (it's only needed on <svg>)
	if el.Tag != "svg" {
		delete(el.Attributes, "xmlns")
	}

	// Remove version (it's optional)
	delete(el.Attributes, "version")
}

// removeWhitespaceAttributes removes whitespace-only attributes.
func removeWhitespaceAttributes(el *models.Element) {
	for key, val := range el.Attributes {
		if strings.TrimSpace(val) == "" {
			delete(el.Attributes, key)
		}
	}
}

// removeEditorAttributes removes editor-specific attributes.
func removeEditorAttributes(el *models.Element) {
	editorAttrs := []string{
		"data-name", "inkscape:version", "sodipodi:docname",
		"inkscape:export-filename", "inkscape:export-xdpi",
		"inkscape:export-ydpi", "sodipodi:cx", "sodipodi:cy",
		"sodipodi:rx", "sodipodi:ry", "sodipodi:type",
		"sodipodi:role", "sodipodi:arc-type",
	}
	for _, attr := range editorAttrs {
		delete(el.Attributes, attr)
	}
}

// simplifyPath simplifies path data.
func simplifyPath(el *models.Element) {
	d := el.Attribute("d")
	if d == "" {
		return
	}
	// Basic simplifications
	d = regexp.MustCompile(`\s+`).ReplaceAllString(d, " ")
	d = strings.TrimSpace(d)
	// Remove trailing zeros after decimal point
	d = regexp.MustCompile(`(\.\d*?)0+(\s|$)`).ReplaceAllString(d, "$1$2")
	// Remove leading zeros in numbers
	d = regexp.MustCompile(` 0(\d)`).ReplaceAllString(d, " $1")
	el.SetAttribute("d", d)
}

// cleanDefs removes empty <defs> elements.
func cleanDefs(doc *models.SVGDocument) {
	result := make([]*models.Element, 0, len(doc.Defs))
	for _, def := range doc.Defs {
		if len(def.Children) > 0 {
			result = append(result, def)
		}
	}
	doc.Defs = result
}

// EstimateSize estimates the byte size of the optimized SVG.
func EstimateSize(doc *models.SVGDocument) int {
	size := 0
	for _, el := range doc.Elements {
		size += estimateElementSize(el)
	}
	return size
}

func estimateElementSize(el *models.Element) int {
	size := len(el.Tag) + 2 // <tag>
	for k, v := range el.Attributes {
		size += len(k) + len(v) + 3 // key="value"
	}
	size += len(el.Text)
	for _, child := range el.Children {
		size += estimateElementSize(child)
	}
	return size
}
