package transform_test

import (
	"math"
	"testing"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/transform"
)

func TestIdentity(t *testing.T) {
	tf := transform.Identity()
	if tf.A != 1 || tf.B != 0 || tf.C != 0 || tf.D != 1 || tf.E != 0 || tf.F != 0 {
		t.Errorf("Identity = %v, want {1, 0, 0, 1, 0, 0}", tf)
	}
}

func TestScale(t *testing.T) {
	tf := transform.Scale(2, 3)
	if tf.A != 2 || tf.D != 3 {
		t.Errorf("Scale(2,3) = %v, want {A:2, D:3}", tf)
	}
}

func TestTranslate(t *testing.T) {
	tf := transform.Translate(10, 20)
	if tf.E != 10 || tf.F != 20 {
		t.Errorf("Translate(10,20) = %v, want {E:10, F:20}", tf)
	}
}

func TestRotate(t *testing.T) {
	tf := transform.Rotate(90)
	// cos(90°) ≈ 0, sin(90°) ≈ 1
	if math.Abs(tf.A) > 0.001 || math.Abs(tf.B-1) > 0.001 || math.Abs(tf.C+1) > 0.001 || math.Abs(tf.D) > 0.001 {
		t.Errorf("Rotate(90) = %v, want approximately {0, 1, -1, 0, 0, 0}", tf)
	}
}

func TestCompose(t *testing.T) {
	s := transform.Scale(2, 2)
	tr := transform.Translate(10, 0)
	result := transform.Compose(s, tr)
	// Compose(Scale, Translate) applies Translate first, then Scale
	// point (5,0) → translate → (15,0) → scale → (30,0)
	x, y := result.TransformPoint(5, 0)
	if math.Abs(x-30) > 0.001 || math.Abs(y-0) > 0.001 {
		t.Errorf("Compose(Scale,Translate).TransformPoint(5,0) = (%.2f, %.2f), want (30, 0)", x, y)
	}
}

func TestTransformPoint(t *testing.T) {
	tf := transform.Translate(10, 20)
	x, y := tf.TransformPoint(5, 5)
	if x != 15 || y != 25 {
		t.Errorf("TransformPoint(5,5) = (%.2f, %.2f), want (15, 25)", x, y)
	}
}

func TestInverse(t *testing.T) {
	tf := transform.Scale(2, 3)
	inv := tf.Inverse()
	// Scale(2,3) * Inverse = Identity
	combined := transform.Compose(tf, inv)
	x, y := combined.TransformPoint(5, 10)
	if math.Abs(x-5) > 0.001 || math.Abs(y-10) > 0.001 {
		t.Errorf("Scale * Inverse should be identity, got TransformPoint(5,10) = (%.2f, %.2f)", x, y)
	}
}

func TestSkewX(t *testing.T) {
	tf := transform.SkewX(45)
	// tan(45°) = 1
	if math.Abs(tf.C-1) > 0.001 {
		t.Errorf("SkewX(45).C = %f, want ~1", tf.C)
	}
}

func TestSkewY(t *testing.T) {
	tf := transform.SkewY(45)
	if math.Abs(tf.B-1) > 0.001 {
		t.Errorf("SkewY(45).B = %f, want ~1", tf.B)
	}
}

func TestToSVGAttribute(t *testing.T) {
	tests := []struct {
		name string
		tf   transform.Transform
		want string
	}{
		{"identity", transform.Identity(), ""},
		{"translate", transform.Translate(10, 20), "translate(10, 20)"},
		{"scale", transform.Scale(2, 2), "scale(2)"},
		{"scale-xy", transform.Scale(2, 3), "scale(2, 3)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tf.ToSVGAttribute()
			if got != tt.want {
				t.Errorf("ToSVGAttribute() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseTransform(t *testing.T) {
	tests := []struct {
		name string
		s    string
	}{
		{"translate", "translate(10, 20)"},
		{"scale", "scale(2)"},
		{"rotate", "rotate(45)"},
		{"matrix", "matrix(1, 0, 0, 1, 10, 20)"},
		{"combined", "translate(10, 0) rotate(45)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf := transform.ParseTransform(tt.s)
			// Just verify it doesn't panic and produces a valid transform
			_, _ = tf.TransformPoint(0, 0)
		})
	}
}

func TestRotateAround(t *testing.T) {
	tf := transform.RotateAround(90, 0, 0)
	// Should be same as Rotate(90) when rotating around origin
	x, y := tf.TransformPoint(1, 0)
	if math.Abs(x-0) > 0.001 || math.Abs(y-1) > 0.001 {
		t.Errorf("RotateAround(90,0,0).TransformPoint(1,0) = (%.2f, %.2f), want (0, 1)", x, y)
	}
}
