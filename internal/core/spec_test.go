package core

import (
	"slices"
	"testing"
)

func TestSpecValid(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		wantErr bool
	}{
		{"ok", Spec{Dim: 2, Low: []float64{-1, 0}, High: []float64{1, 1}}, false},
		{"zero dim", Spec{Dim: 0}, true},
		{"low len mismatch", Spec{Dim: 2, Low: []float64{0}, High: []float64{1, 1}}, true},
		{"high len mismatch", Spec{Dim: 2, Low: []float64{0, 0}, High: []float64{1}}, true},
		{"low above high", Spec{Dim: 1, Low: []float64{2}, High: []float64{1}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.spec.Valid(); (err != nil) != tt.wantErr {
				t.Fatalf("Valid() err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestSpecClamp(t *testing.T) {
	s := Spec{Dim: 2, Low: []float64{-1, 0}, High: []float64{1, 1}}
	got := s.Clamp([]float64{-5, 2})
	if want := []float64{-1, 1}; !slices.Equal(got, want) {
		t.Fatalf("Clamp = %v, want %v", got, want)
	}
}
