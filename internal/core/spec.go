package core

import "fmt"

// Valid reports whether the spec is internally consistent: a positive dimension
// and bound slices whose lengths match that dimension.
func (s Spec) Valid() error {
	if s.Dim <= 0 {
		return fmt.Errorf("spec: dim must be positive, got %d", s.Dim)
	}
	if len(s.Low) != s.Dim {
		return fmt.Errorf("spec: len(Low)=%d, want %d", len(s.Low), s.Dim)
	}
	if len(s.High) != s.Dim {
		return fmt.Errorf("spec: len(High)=%d, want %d", len(s.High), s.Dim)
	}
	for i := range s.Dim {
		if s.Low[i] > s.High[i] {
			return fmt.Errorf("spec: component %d has Low %g > High %g", i, s.Low[i], s.High[i])
		}
	}
	return nil
}

// Clamp returns v with each component constrained to the spec's bounds.
func (s Spec) Clamp(v []float64) []float64 {
	out := make([]float64, len(v))
	for i := range v {
		if i < s.Dim {
			out[i] = min(max(v[i], s.Low[i]), s.High[i])
		} else {
			out[i] = v[i]
		}
	}
	return out
}
