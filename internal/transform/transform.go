// Package transform provides SVG transformation operations.
package transform

import (
	"fmt"
	"math"
	"strings"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
)

// Transform represents an affine transformation.
type Transform struct {
	A, B, C, D, E, F float64
}

// Identity returns the identity transformation.
func Identity() Transform {
	return Transform{A: 1, D: 1}
}

// Scale creates a scale transformation.
func Scale(sx, sy float64) Transform {
	return Transform{A: sx, D: sy}
}

// Translate creates a translation transformation.
func Translate(tx, ty float64) Transform {
	return Transform{A: 1, D: 1, E: tx, F: ty}
}

// Rotate creates a rotation transformation (angle in degrees).
func Rotate(angle float64) Transform {
	rad := angle * math.Pi / 180
	cos := math.Cos(rad)
	sin := math.Sin(rad)
	return Transform{A: cos, B: sin, C: -sin, D: cos}
}

// RotateAround creates a rotation around a point.
func RotateAround(angle, cx, cy float64) Transform {
	return Compose(
		Compose(Translate(-cx, -cy), Rotate(angle)),
		Translate(cx, cy),
	)
}

// SkewX creates a horizontal skew transformation.
func SkewX(angle float64) Transform {
	rad := angle * math.Pi / 180
	return Transform{A: 1, C: math.Tan(rad), D: 1}
}

// SkewY creates a vertical skew transformation.
func SkewY(angle float64) Transform {
	rad := angle * math.Pi / 180
	return Transform{A: 1, B: math.Tan(rad), D: 1}
}

// Compose multiplies two transformations.
func Compose(t1, t2 Transform) Transform {
	return Transform{
		A: t1.A*t2.A + t1.C*t2.B,
		B: t1.B*t2.A + t1.D*t2.B,
		C: t1.A*t2.C + t1.C*t2.D,
		D: t1.B*t2.C + t1.D*t2.D,
		E: t1.A*t2.E + t1.C*t2.F + t1.E,
		F: t1.B*t2.E + t1.D*t2.F + t1.F,
	}
}

// Inverse returns the inverse transformation.
func (t Transform) Inverse() Transform {
	det := t.A*t.D - t.B*t.C
	if math.Abs(det) < 1e-10 {
		return Identity() // singular matrix
	}
	return Transform{
		A: t.D / det,
		B: -t.B / det,
		C: -t.C / det,
		D: t.A / det,
		E: (t.C*t.F - t.D*t.E) / det,
		F: (t.B*t.E - t.A*t.F) / det,
	}
}

// TransformPoint applies the transformation to a point.
func (t Transform) TransformPoint(x, y float64) (float64, float64) {
	newX := t.A*x + t.C*y + t.E
	newY := t.B*x + t.D*y + t.F
	return newX, newY
}

// ToSVGAttribute returns the transform as an SVG transform attribute string.
func (t Transform) ToSVGAttribute() string {
	// Check for simple transformations
	if t.A == 1 && t.B == 0 && t.C == 0 && t.D == 1 {
		if t.E != 0 || t.F != 0 {
			return fmt.Sprintf("translate(%.4g, %.4g)", t.E, t.F)
		}
		return ""
	}

	if t.E == 0 && t.F == 0 && t.B == 0 && t.C == 0 {
		if t.A == t.D {
			return fmt.Sprintf("scale(%.4g)", t.A)
		}
		return fmt.Sprintf("scale(%.4g, %.4g)", t.A, t.D)
	}

	// General matrix
	return fmt.Sprintf("matrix(%.4g, %.4g, %.4g, %.4g, %.4g, %.4g)",
		t.A, t.B, t.C, t.D, t.E, t.F)
}

// ParseTransform parses an SVG transform attribute string.
func ParseTransform(s string) Transform {
	t := Identity()
	s = strings.TrimSpace(s)

	for s != "" {
		// Find function name
		idx := strings.Index(s, "(")
		if idx < 0 {
			break
		}
		funcName := strings.TrimSpace(s[:idx])
		s = s[idx+1:]

		// Find closing paren
		endIdx := strings.Index(s, ")")
		if endIdx < 0 {
			break
		}
		argsStr := strings.TrimSpace(s[:endIdx])
		s = strings.TrimSpace(s[endIdx+1:])

		// Parse args
		args := parseArgs(argsStr)

		var local Transform
		switch strings.ToLower(funcName) {
		case "translate":
			if len(args) >= 2 {
				local = Translate(args[0], args[1])
			} else if len(args) == 1 {
				local = Translate(args[0], 0)
			}
		case "scale":
			if len(args) >= 2 {
				local = Scale(args[0], args[1])
			} else if len(args) == 1 {
				local = Scale(args[0], args[0])
			}
		case "rotate":
			if len(args) >= 3 {
				local = RotateAround(args[0], args[1], args[2])
			} else if len(args) == 1 {
				local = Rotate(args[0])
			}
		case "skewx":
			if len(args) >= 1 {
				local = SkewX(args[0])
			}
		case "skewy":
			if len(args) >= 1 {
				local = SkewY(args[0])
			}
		case "matrix":
			if len(args) >= 6 {
				local = Transform{
					A: args[0], B: args[1], C: args[2],
					D: args[3], E: args[4], F: args[5],
				}
			}
		}

		t = Compose(t, local)
	}

	return t
}

func parseArgs(s string) []float64 {
	var args []float64
	s = strings.TrimSpace(s)
	for s != "" {
		var f float64
		n, err := fmt.Sscanf(s, "%f", &f)
		if err != nil || n == 0 {
			break
		}
		args = append(args, f)
		// Skip past the number
		for i := 0; i < len(s); i++ {
			if s[i] == ',' || s[i] == ' ' || s[i] == '\t' || s[i] == '\n' {
				s = strings.TrimSpace(s[i+1:])
				break
			}
			if i == len(s)-1 {
				s = ""
			}
		}
	}
	return args
}

// ScaleElement scales an element by the given factors.
func ScaleElement(el *models.Element, sx, sy float64) {
	t := ParseTransform(el.Transform())
	t = Compose(t, Scale(sx, sy))
	attr := t.ToSVGAttribute()
	if attr == "" {
		delete(el.Attributes, "transform")
	} else {
		el.SetAttribute("transform", attr)
	}
}

// TranslateElement translates an element.
func TranslateElement(el *models.Element, tx, ty float64) {
	t := ParseTransform(el.Transform())
	t = Compose(t, Translate(tx, ty))
	attr := t.ToSVGAttribute()
	if attr == "" {
		delete(el.Attributes, "transform")
	} else {
		el.SetAttribute("transform", attr)
	}
}

// RotateElement rotates an element by the given angle (degrees).
func RotateElement(el *models.Element, angle float64) {
	t := ParseTransform(el.Transform())
	t = Compose(t, Rotate(angle))
	attr := t.ToSVGAttribute()
	if attr == "" {
		delete(el.Attributes, "transform")
	} else {
		el.SetAttribute("transform", attr)
	}
}
