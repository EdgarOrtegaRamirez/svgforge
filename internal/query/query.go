// Package query provides CSS-like selector matching for SVG elements.
package query

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
)

// Selector represents a parsed CSS-like selector.
type Selector struct {
	Parts []SelectorPart
}

// SelectorPart represents one part of a selector chain.
type SelectorPart struct {
	Combinator string // "", " " (descendant), ">" (child), "+" (sibling), "~" (general sibling)
	Tag        string
	ID         string
	Classes    []string
	Attributes []AttrCondition
	Pseudo     string
}

// AttrCondition represents an attribute condition.
type AttrCondition struct {
	Name     string
	Operator string // "=", "~=", "|=", "^=", "$=", "*="
	Value    string
}

// Parse parses a CSS-like selector string.
func Parse(selector string) (*Selector, error) {
	s := &Selector{}
	selector = strings.TrimSpace(selector)

	for selector != "" {
		part, remaining, err := parsePart(selector)
		if err != nil {
			return nil, err
		}
		s.Parts = append(s.Parts, part)
		selector = remaining
	}

	return s, nil
}

func parsePart(s string) (SelectorPart, string, error) {
	part := SelectorPart{}
	remaining := strings.TrimSpace(s)

	// Parse combinator
	if len(remaining) > 0 {
		switch remaining[0] {
		case '>':
			part.Combinator = ">"
			remaining = strings.TrimSpace(remaining[1:])
		case '+':
			part.Combinator = "+"
			remaining = strings.TrimSpace(remaining[1:])
		case '~':
			part.Combinator = "~"
			remaining = strings.TrimSpace(remaining[1:])
		default:
			// If there's a previous part, this is a descendant combinator
			if part.Combinator == "" {
				part.Combinator = " "
			}
		}
	}

	// Parse tag
	if len(remaining) > 0 && remaining[0] != '#' && remaining[0] != '.' && remaining[0] != '[' && remaining[0] != ':' {
		i := 0
		for i < len(remaining) && remaining[i] != ' ' && remaining[i] != '#' && remaining[i] != '.' && remaining[i] != '[' && remaining[i] != ':' && remaining[i] != '>' && remaining[i] != '+' && remaining[i] != '~' {
			i++
		}
		part.Tag = remaining[:i]
		remaining = strings.TrimSpace(remaining[i:])
	}

	// Parse ID, classes, and attributes in any order
	for len(remaining) > 0 {
		switch remaining[0] {
		case '#':
			remaining = remaining[1:]
			i := 0
			for i < len(remaining) && remaining[i] != ' ' && remaining[i] != '.' && remaining[i] != '#' && remaining[i] != '[' && remaining[i] != ':' && remaining[i] != '>' && remaining[i] != '+' && remaining[i] != '~' {
				i++
			}
			part.ID = remaining[:i]
			remaining = strings.TrimSpace(remaining[i:])
		case '.':
			remaining = remaining[1:]
			i := 0
			for i < len(remaining) && remaining[i] != ' ' && remaining[i] != '.' && remaining[i] != '#' && remaining[i] != '[' && remaining[i] != ':' && remaining[i] != '>' && remaining[i] != '+' && remaining[i] != '~' {
				i++
			}
			part.Classes = append(part.Classes, remaining[:i])
			remaining = strings.TrimSpace(remaining[i:])
		case '[':
			remaining = remaining[1:]
			i := strings.Index(remaining, "]")
			if i < 0 {
				return part, "", fmt.Errorf("unclosed attribute selector")
			}
			attrStr := strings.TrimSpace(remaining[:i])
			remaining = strings.TrimSpace(remaining[i+1:])
			cond := parseAttrCondition(attrStr)
			part.Attributes = append(part.Attributes, cond)
		case ':':
			remaining = remaining[1:]
			i := 0
			for i < len(remaining) && remaining[i] != ' ' && remaining[i] != '>' && remaining[i] != '+' && remaining[i] != '~' {
				i++
			}
			part.Pseudo = remaining[:i]
			remaining = strings.TrimSpace(remaining[i:])
		default:
			// Not a selector part character, stop
			goto done
		}
	}
done:

	return part, remaining, nil
}

func parseAttrCondition(s string) AttrCondition {
	cond := AttrCondition{}
	ops := []string{"~=", "|=", "^=", "$=", "*=", "="}
	for _, op := range ops {
		if idx := strings.Index(s, op); idx >= 0 {
			cond.Name = strings.TrimSpace(s[:idx])
			cond.Operator = op
			val := strings.TrimSpace(s[idx+len(op):])
			// Remove quotes
			if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') {
				val = val[1 : len(val)-1]
			}
			cond.Value = val
			return cond
		}
	}
	cond.Name = strings.TrimSpace(s)
	cond.Operator = ""
	return cond
}

// Match checks if an element matches the selector.
func Match(el *models.Element, selector *Selector) bool {
	if len(selector.Parts) == 0 {
		return false
	}
	return matchFromEnd(el, selector.Parts, len(selector.Parts)-1)
}

func matchFromEnd(el *models.Element, parts []SelectorPart, idx int) bool {
	if idx < 0 {
		return el != nil
	}
	if el == nil {
		return false
	}

	part := parts[idx]
	if !matchesPart(el, part) {
		return false
	}

	if idx == 0 {
		return true
	}

	// Find matching ancestors/siblings based on combinator
	switch part.Combinator {
	case " ":
		// Descendant: match any ancestor
		parent := el.Parent
		for parent != nil {
			if matchFromEnd(parent, parts, idx-1) {
				return true
			}
			parent = parent.Parent
		}
		return false
	case ">":
		// Child: match direct parent
		return matchFromEnd(el.Parent, parts, idx-1)
	case "+":
		// Adjacent sibling
		sibling := findPreviousSibling(el)
		return matchFromEnd(sibling, parts, idx-1)
	case "~":
		// General sibling
		sibling := el.Parent
		if sibling == nil {
			return false
		}
		for _, child := range sibling.Children {
			if child == el {
				break
			}
			if matchFromEnd(child, parts, idx-1) {
				return true
			}
		}
		return false
	}

	return false
}

func matchesPart(el *models.Element, part SelectorPart) bool {
	// Match tag
	if part.Tag != "" && part.Tag != "*" && el.Tag != part.Tag {
		return false
	}

	// Match ID
	if part.ID != "" && el.ID() != part.ID {
		return false
	}

	// Match classes
	for _, class := range part.Classes {
		if !el.HasClass(class) {
			return false
		}
	}

	// Match attributes
	for _, cond := range part.Attributes {
		if !matchesAttr(el, cond) {
			return false
		}
	}

	// Match pseudo-class
	if part.Pseudo != "" {
		if !matchesPseudo(el, part.Pseudo) {
			return false
		}
	}

	return true
}

func matchesAttr(el *models.Element, cond AttrCondition) bool {
	val := el.Attribute(cond.Name)
	switch cond.Operator {
	case "":
		return val != "" // presence check
	case "=":
		return val == cond.Value
	case "~=":
		return strings.Contains(" "+val+" ", " "+cond.Value+" ")
	case "|=":
		return val == cond.Value || strings.HasPrefix(val, cond.Value+"-")
	case "^=":
		return strings.HasPrefix(val, cond.Value)
	case "$=":
		return strings.HasSuffix(val, cond.Value)
	case "*=":
		return strings.Contains(val, cond.Value)
	}
	return false
}

func matchesPseudo(el *models.Element, pseudo string) bool {
	switch pseudo {
	case "first-child":
		return el.Parent != nil && len(el.Parent.Children) > 0 && el.Parent.Children[0] == el
	case "last-child":
		return el.Parent != nil && len(el.Parent.Children) > 0 && el.Parent.Children[len(el.Parent.Children)-1] == el
	case "empty":
		return len(el.Children) == 0 && el.Text == ""
	case "root":
		return el.Tag == "svg"
	case "not-empty":
		return len(el.Children) > 0 || el.Text != ""
	default:
		// Handle :nth-child(n)
		if strings.HasPrefix(pseudo, "nth-child(") && strings.HasSuffix(pseudo, ")") {
			n := pseudo[len("nth-child(") : len(pseudo)-1]
			idx, err := strconv.Atoi(n)
			if err != nil {
				return false
			}
			if el.Parent == nil || idx < 1 || idx > len(el.Parent.Children) {
				return false
			}
			return el.Parent.Children[idx-1] == el
		}
		return false
	}
}

func findPreviousSibling(el *models.Element) *models.Element {
	if el.Parent == nil {
		return nil
	}
	for i, child := range el.Parent.Children {
		if child == el && i > 0 {
			return el.Parent.Children[i-1]
		}
	}
	return nil
}

// Query finds all elements matching a selector.
func Query(root *models.Element, selector *Selector) []*models.Element {
	var results []*models.Element
	findMatches(root, selector, &results)
	return results
}

func findMatches(el *models.Element, selector *Selector, results *[]*models.Element) {
	if Match(el, selector) {
		*results = append(*results, el)
	}
	for _, child := range el.Children {
		findMatches(child, selector, results)
	}
}

// QueryString is a convenience function that parses a selector string and queries.
func QueryString(root *models.Element, selectorStr string) ([]*models.Element, error) {
	selector, err := Parse(selectorStr)
	if err != nil {
		return nil, err
	}
	return Query(root, selector), nil
}
