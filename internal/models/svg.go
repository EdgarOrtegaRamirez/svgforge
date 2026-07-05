// Package models defines the SVG AST data structures.
package models

import (
	"fmt"
	"strings"
)

// SVGDocument represents a parsed SVG document.
type SVGDocument struct {
	Version  string
	Width    string
	Height   string
	ViewBox  *ViewBox
	Xmlns    string
	Title    string
	Desc     string
	Defs     []*Element
	Elements []*Element
	Styles   []StyleRule
	Metadata map[string]string
}

// ViewBox represents the SVG viewBox attribute.
type ViewBox struct {
	MinX   float64
	MinY   float64
	Width  float64
	Height float64
}

// Element represents an SVG element.
type Element struct {
	Tag        string
	Attributes map[string]string
	Children   []*Element
	Text       string
	Parent     *Element
	Line       int
	Column     int
}

// StyleRule represents a CSS style rule.
type StyleRule struct {
	Selector string
	Properties map[string]string
}

// Attribute returns an attribute value, empty string if not found.
func (e *Element) Attribute(name string) string {
	if e.Attributes == nil {
		return ""
	}
	return e.Attributes[name]
}

// SetAttribute sets an attribute on the element.
func (e *Element) SetAttribute(name, value string) {
	if e.Attributes == nil {
		e.Attributes = make(map[string]string)
	}
	e.Attributes[name] = value
}

// HasClass checks if the element has a specific CSS class.
func (e *Element) HasClass(class string) bool {
	classes := strings.Fields(e.Attribute("class"))
	for _, c := range classes {
		if c == class {
			return true
		}
	}
	return false
}

// AddClass adds a CSS class to the element.
func (e *Element) AddClass(class string) {
	existing := e.Attribute("class")
	if existing == "" {
		e.SetAttribute("class", class)
		return
	}
	classes := strings.Fields(existing)
	for _, c := range classes {
		if c == class {
			return // already has class
		}
	}
	e.SetAttribute("class", strings.Join(append(classes, class), " "))
}

// RemoveClass removes a CSS class from the element.
func (e *Element) RemoveClass(class string) {
	classes := strings.Fields(e.Attribute("class"))
	result := make([]string, 0, len(classes))
	for _, c := range classes {
		if c != class {
			result = append(result, c)
		}
	}
	if len(result) == 0 {
		delete(e.Attributes, "class")
	} else {
		e.SetAttribute("class", strings.Join(result, " "))
	}
}

// ID returns the element's id attribute.
func (e *Element) ID() string {
	return e.Attribute("id")
}

// Transform returns the element's transform attribute.
func (e *Element) Transform() string {
	return e.Attribute("transform")
}

// BBox returns the bounding box from attributes (x, y, width, height).
func (e *Element) BBox() (x, y, w, h float64) {
	x = parseFloat(e.Attribute("x"))
	y = parseFloat(e.Attribute("y"))
	w = parseFloat(e.Attribute("width"))
	h = parseFloat(e.Attribute("height"))
	return
}

// IsContainer checks if the element can contain children.
func (e *Element) IsContainer() bool {
	switch e.Tag {
	case "g", "svg", "defs", "symbol", "clipPath", "mask", "pattern", "marker", "a", "switch":
		return true
	}
	return false
}

// IsShape checks if the element is a shape element.
func (e *Element) IsShape() bool {
	switch e.Tag {
	case "rect", "circle", "ellipse", "line", "polyline", "polygon", "path", "text", "image", "use":
		return true
	}
	return false
}

// FindElements recursively finds all elements matching a predicate.
func (e *Element) FindElements(predicate func(*Element) bool) []*Element {
	var result []*Element
	if predicate(e) {
		result = append(result, e)
	}
	for _, child := range e.Children {
		result = append(result, child.FindElements(predicate)...)
	}
	return result
}

// FindByTag finds all descendant elements with the given tag.
func (e *Element) FindByTag(tag string) []*Element {
	return e.FindElements(func(el *Element) bool {
		return el.Tag == tag
	})
}

// FindByID finds the first descendant element with the given id.
func (e *Element) FindByID(id string) *Element {
	results := e.FindElements(func(el *Element) bool {
		return el.ID() == id
	})
	if len(results) > 0 {
		return results[0]
	}
	return nil
}

// FindByClass finds all descendants with the given CSS class.
func (e *Element) FindByClass(class string) []*Element {
	return e.FindElements(func(el *Element) bool {
		return el.HasClass(class)
	})
}

// CountElements counts all descendant elements (including self).
func (e *Element) CountElements() int {
	count := 1
	for _, child := range e.Children {
		count += child.CountElements()
	}
	return count
}

// Depth returns the maximum depth of the element tree.
func (e *Element) Depth() int {
	maxChild := 0
	for _, child := range e.Children {
		if d := child.Depth(); d > maxChild {
			maxChild = d
		}
	}
	return maxChild + 1
}

// RemoveChild removes a child element.
func (e *Element) RemoveChild(child *Element) bool {
	for i, c := range e.Children {
		if c == child {
			e.Children = append(e.Children[:i], e.Children[i+1:]...)
			return true
		}
	}
	return false
}

// AddChild adds a child element.
func (e *Element) AddChild(child *Element) {
	child.Parent = e
	e.Children = append(e.Children, child)
}

// InsertBefore inserts a child before the reference element.
func (e *Element) InsertBefore(ref, child *Element) bool {
	for i, c := range e.Children {
		if c == ref {
			child.Parent = e
			e.Children = append(e.Children[:i+1], e.Children[i:]...)
			e.Children[i] = child
			return true
		}
	}
	return false
}

// String returns a short description of the element.
func (e *Element) String() string {
	if e.Text != "" {
		text := e.Text
		if len(text) > 30 {
			text = text[:30] + "..."
		}
		return fmt.Sprintf("<%s>%s</%s>", e.Tag, text, e.Tag)
	}
	return fmt.Sprintf("<%s>", e.Tag)
}

// ElementStats holds statistics about an SVG document.
type ElementStats struct {
	TotalElements  int
	ShapeElements  int
	ContainerElements int
	TextElements   int
	UseElements    int
	DefElements    int
	MaxDepth       int
	TagCounts      map[string]int
	ClassCounts    map[string]int
	IDCounts       map[string]int
}

// parseFloat is a helper that parses a float from string, returning 0 on error.
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
