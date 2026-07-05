// Package diff provides structural SVG diffing.
package diff

import (
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
)

// ChangeType represents the type of change.
type ChangeType int

const (
	ChangeAdded ChangeType = iota
	ChangeRemoved
	ChangeModified
	ChangeUnchanged
)

// DiffEntry represents a single change in a diff.
type DiffEntry struct {
	Type     ChangeType
	Path     string
	OldValue string
	NewValue string
	Element  *models.Element
}

// DiffResult holds the result of comparing two SVG documents.
type DiffResult struct {
	Entries  []DiffEntry
	Summary  DiffSummary
}

// DiffSummary provides a summary of changes.
type DiffSummary struct {
	Added    int
	Removed  int
	Modified int
	Unchanged int
}

// Diff compares two SVG documents and returns the differences.
func Diff(doc1, doc2 *models.SVGDocument) *DiffResult {
	result := &DiffResult{
		Entries: make([]DiffEntry, 0),
	}

	// Compare root attributes
	compareRootAttrs(doc1, doc2, result)

	// Compare elements
	compareElements(doc1.Elements, doc2.Elements, "", result)

	// Calculate summary
	for _, entry := range result.Entries {
		switch entry.Type {
		case ChangeAdded:
			result.Summary.Added++
		case ChangeRemoved:
			result.Summary.Removed++
		case ChangeModified:
			result.Summary.Modified++
		case ChangeUnchanged:
			result.Summary.Unchanged++
		}
	}

	return result
}

func compareRootAttrs(doc1, doc2 *models.SVGDocument, result *DiffResult) {
	if doc1.Width != doc2.Width {
		result.Entries = append(result.Entries, DiffEntry{
			Type:     ChangeModified,
			Path:     "svg[@width]",
			OldValue: doc1.Width,
			NewValue: doc2.Width,
		})
	}
	if doc1.Height != doc2.Height {
		result.Entries = append(result.Entries, DiffEntry{
			Type:     ChangeModified,
			Path:     "svg[@height]",
			OldValue: doc1.Height,
			NewValue: doc2.Height,
		})
	}
	if doc1.Title != doc2.Title {
		result.Entries = append(result.Entries, DiffEntry{
			Type:     ChangeModified,
			Path:     "svg/title",
			OldValue: doc1.Title,
			NewValue: doc2.Title,
		})
	}
}

func compareElements(elems1, elems2 []*models.Element, basePath string, result *DiffResult) {
	// Match elements by tag+id and then by position
	used1 := make([]bool, len(elems1))
	used2 := make([]bool, len(elems2))

	// First pass: match by tag+id
	for i, e1 := range elems1 {
		if used1[i] {
			continue
		}
		for j, e2 := range elems2 {
			if used2[j] {
				continue
			}
			if e1.Tag == e2.Tag && e1.ID() != "" && e1.ID() == e2.ID() {
				path := fmt.Sprintf("%s/%s[@id='%s']", basePath, e1.Tag, e1.ID())
				compareElement(e1, e2, path, result)
				used1[i] = true
				used2[j] = true
				break
			}
		}
	}

	// Second pass: match by tag+index
	for i, e1 := range elems1 {
		if used1[i] {
			continue
		}
		for j, e2 := range elems2 {
			if used2[j] {
				continue
			}
			if e1.Tag == e2.Tag {
				path := fmt.Sprintf("%s/%s[%d]", basePath, e1.Tag, j)
				compareElement(e1, e2, path, result)
				used1[i] = true
				used2[j] = true
				break
			}
		}
	}

	// Report removed elements
	for i, e1 := range elems1 {
		if !used1[i] {
			path := fmt.Sprintf("%s/%s", basePath, e1.Tag)
			result.Entries = append(result.Entries, DiffEntry{
				Type:    ChangeRemoved,
				Path:    path,
				Element: e1,
			})
		}
	}

	// Report added elements
	for j, e2 := range elems2 {
		if !used2[j] {
			path := fmt.Sprintf("%s/%s", basePath, e2.Tag)
			result.Entries = append(result.Entries, DiffEntry{
				Type:    ChangeAdded,
				Path:    path,
				Element: e2,
			})
		}
	}
}

func compareElement(e1, e2 *models.Element, path string, result *DiffResult) {
	// Compare attributes
	changed := false
	allKeys := make(map[string]bool)
	for k := range e1.Attributes {
		allKeys[k] = true
	}
	for k := range e2.Attributes {
		allKeys[k] = true
	}

	for k := range allKeys {
		v1 := e1.Attribute(k)
		v2 := e2.Attribute(k)
		if v1 != v2 {
			changed = true
			result.Entries = append(result.Entries, DiffEntry{
				Type:     ChangeModified,
				Path:     fmt.Sprintf("%s/@%s", path, k),
				OldValue: v1,
				NewValue: v2,
			})
		}
	}

	// Compare text
	if e1.Text != e2.Text {
		changed = true
		result.Entries = append(result.Entries, DiffEntry{
			Type:     ChangeModified,
			Path:     path + "/text()",
			OldValue: e1.Text,
			NewValue: e2.Text,
		})
	}

	// Compare children
	compareElements(e1.Children, e2.Children, path, result)

	if !changed && len(e1.Children) == 0 && len(e2.Children) == 0 {
		result.Entries = append(result.Entries, DiffEntry{
			Type: ChangeUnchanged,
			Path: path,
		})
	}
}

// FormatText formats a diff result as text.
func FormatText(result *DiffResult) string {
	var sb strings.Builder

	for _, entry := range result.Entries {
		switch entry.Type {
		case ChangeAdded:
			sb.WriteString(fmt.Sprintf("+ %s\n", entry.Path))
		case ChangeRemoved:
			sb.WriteString(fmt.Sprintf("- %s\n", entry.Path))
		case ChangeModified:
			sb.WriteString(fmt.Sprintf("~ %s: %q -> %q\n", entry.Path, entry.OldValue, entry.NewValue))
		case ChangeUnchanged:
			sb.WriteString(fmt.Sprintf("  %s\n", entry.Path))
		}
	}

	sb.WriteString(fmt.Sprintf("\nSummary: %d added, %d removed, %d modified, %d unchanged\n",
		result.Summary.Added, result.Summary.Removed, result.Summary.Modified, result.Summary.Unchanged))

	return sb.String()
}

// FormatCompact formats a diff result compactly.
func FormatCompact(result *DiffResult) string {
	return fmt.Sprintf("+%d -%d ~%d",
		result.Summary.Added, result.Summary.Removed, result.Summary.Modified)
}
