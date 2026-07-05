// Package stats provides SVG statistics and analysis.
package stats

import (
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
)

// Stats holds comprehensive SVG statistics.
type Stats struct {
	Width           string            `json:"width"`
	Height          string            `json:"height"`
	ViewBox         string            `json:"view_box,omitempty"`
	TotalElements   int               `json:"total_elements"`
	ShapeElements   int               `json:"shape_elements"`
	ContainerElements int             `json:"container_elements"`
	TextElements    int               `json:"text_elements"`
	UseElements     int               `json:"use_elements"`
	DefElements     int               `json:"def_elements"`
	MaxDepth        int               `json:"max_depth"`
	TagCounts       map[string]int    `json:"tag_counts"`
	ClassCounts     map[string]int    `json:"class_counts"`
	IDCounts        map[string]int    `json:"id_counts"`
	AttrCounts      map[string]int    `json:"attr_counts"`
	HasStyles       bool              `json:"has_styles"`
	HasScripts      bool              `json:"has_scripts"`
	HasAnimations   bool              `json:"has_animations"`
	HasFilters      bool              `json:"has_filters"`
	HasGradients    bool              `json:"has_gradients"`
	HasPatterns     bool              `json:"has_patterns"`
	HasMasks        bool              `json:"has_masks"`
	HasClipPaths    bool              `json:"has_clip_paths"`
	EstimatedSize   int               `json:"estimated_size"`
}

// Analyze computes statistics for an SVG document.
func Analyze(doc *models.SVGDocument) *Stats {
	s := &Stats{
		Width:     doc.Width,
		Height:    doc.Height,
		TagCounts: make(map[string]int),
		ClassCounts: make(map[string]int),
		IDCounts:  make(map[string]int),
		AttrCounts: make(map[string]int),
	}

	if doc.ViewBox != nil {
		s.ViewBox = fmt.Sprintf("%.4g %.4g %.4g %.4g",
			doc.ViewBox.MinX, doc.ViewBox.MinY, doc.ViewBox.Width, doc.ViewBox.Height)
	}

	for _, el := range doc.Elements {
		analyzeElement(el, s, 0)
	}

	s.EstimatedSize = estimateDocSize(doc)

	return s
}

func analyzeElement(el *models.Element, s *Stats, depth int) {
	if el == nil || el.Tag == "" {
		return
	}

	s.TotalElements++
	if depth+1 > s.MaxDepth {
		s.MaxDepth = depth + 1
	}

	// Count tags
	s.TagCounts[el.Tag]++

	// Count classes
	if class := el.Attribute("class"); class != "" {
		for _, c := range strings.Fields(class) {
			s.ClassCounts[c]++
		}
	}

	// Count IDs
	if id := el.ID(); id != "" {
		s.IDCounts[id]++
	}

	// Count attributes
	for attr := range el.Attributes {
		s.AttrCounts[attr]++
	}

	// Categorize elements
	switch el.Tag {
	case "rect", "circle", "ellipse", "line", "polyline", "polygon", "path":
		s.ShapeElements++
	case "g", "svg", "defs", "symbol", "pattern", "marker", "a":
		s.ContainerElements++
	case "text", "tspan", "textPath":
		s.TextElements++
	case "use":
		s.UseElements++
	case "style":
		s.HasStyles = true
	case "script":
		s.HasScripts = true
	case "animate", "animateTransform", "animateMotion", "set":
		s.HasAnimations = true
	case "filter", "feGaussianBlur", "feColorMatrix", "feBlend", "feComposite",
		"feFlood", "feOffset", "feMerge", "feMergeNode", "feMorphology",
		"feTurbulence", "feDisplacementMap", "feDiffuseLighting", "feSpecularLighting":
		s.HasFilters = true
	case "linearGradient", "radialGradient":
		s.HasGradients = true
	case "mask":
		s.HasMasks = true
	case "clipPath":
		s.HasClipPaths = true
	}

	// Recurse
	for _, child := range el.Children {
		analyzeElement(child, s, depth+1)
	}
}

func estimateDocSize(doc *models.SVGDocument) int {
	size := len("<svg></svg>")
	if doc.Title != "" {
		size += len(doc.Title) + 15
	}
	for _, el := range doc.Elements {
		size += estimateElementSize(el)
	}
	return size
}

func estimateElementSize(el *models.Element) int {
	if el == nil {
		return 0
	}
	size := len(el.Tag) + 3 // <tag>
	for k, v := range el.Attributes {
		size += len(k) + len(v) + 4 // key="value"
	}
	size += len(el.Text) + len(el.Tag) + 3 // text + closing tag
	for _, child := range el.Children {
		size += estimateElementSize(child)
	}
	return size
}

// FormatText formats stats as human-readable text.
func FormatText(s *Stats) string {
	var sb strings.Builder
	sb.WriteString("SVG Statistics\n")
	sb.WriteString("==============\n\n")
	sb.WriteString(fmt.Sprintf("Dimensions: %s × %s\n", s.Width, s.Height))
	if s.ViewBox != "" {
		sb.WriteString(fmt.Sprintf("ViewBox:    %s\n", s.ViewBox))
	}
	sb.WriteString(fmt.Sprintf("\nElements:   %d\n", s.TotalElements))
	sb.WriteString(fmt.Sprintf("  Shapes:   %d\n", s.ShapeElements))
	sb.WriteString(fmt.Sprintf("  Containers: %d\n", s.ContainerElements))
	sb.WriteString(fmt.Sprintf("  Text:     %d\n", s.TextElements))
	sb.WriteString(fmt.Sprintf("  Use:      %d\n", s.UseElements))
	sb.WriteString(fmt.Sprintf("  Defs:     %d\n", s.DefElements))
	sb.WriteString(fmt.Sprintf("  Max depth: %d\n", s.MaxDepth))

	if len(s.TagCounts) > 0 {
		sb.WriteString("\nTag distribution:\n")
		for tag, count := range s.TagCounts {
			sb.WriteString(fmt.Sprintf("  %-20s %d\n", tag, count))
		}
	}

	if len(s.ClassCounts) > 0 {
		sb.WriteString("\nClasses:\n")
		for class, count := range s.ClassCounts {
			sb.WriteString(fmt.Sprintf("  %-20s %d\n", class, count))
		}
	}

	features := make([]string, 0)
	if s.HasStyles {
		features = append(features, "styles")
	}
	if s.HasScripts {
		features = append(features, "scripts")
	}
	if s.HasAnimations {
		features = append(features, "animations")
	}
	if s.HasFilters {
		features = append(features, "filters")
	}
	if s.HasGradients {
		features = append(features, "gradients")
	}
	if s.HasPatterns {
		features = append(features, "patterns")
	}
	if s.HasMasks {
		features = append(features, "masks")
	}
	if s.HasClipPaths {
		features = append(features, "clip-paths")
	}

	if len(features) > 0 {
		sb.WriteString("\nFeatures: ")
		sb.WriteString(strings.Join(features, ", "))
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("\nEstimated size: %d bytes\n", s.EstimatedSize))

	return sb.String()
}
