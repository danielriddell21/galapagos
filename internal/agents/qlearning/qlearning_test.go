package qlearning

import (
	"testing"

	"github.com/danielriddell21/galapagos/internal/core"
)

type obsState []float64

func (s obsState) Observation() []float64 { return s }

func TestKeyStableAndBounded(t *testing.T) {
	a := New(Config{Bins: 4, Actions: 2})
	k1 := a.key([]float64{0.1, 0.9})
	k2 := a.key([]float64{0.1, 0.9})
	if k1 != k2 {
		t.Fatal("same observation produced different keys")
	}
	// Out-of-range values clamp into the valid bin range.
	if a.key([]float64{2, -1}) != a.key([]float64{1, 0}) {
		t.Fatal("out-of-range observation not clamped")
	}
}

func TestArgmax(t *testing.T) {
	if got := argmax([]float64{0.1, 0.5, 0.2}); got != 1 {
		t.Fatalf("argmax = %d, want 1", got)
	}
	if got := argmax([]float64{0.3, 0.3}); got != 0 {
		t.Fatalf("argmax tie = %d, want 0", got)
	}
}

func TestTemporalDifferenceUpdate(t *testing.T) {
	a := New(Config{Bins: 2, Actions: 2, Alpha: 0.5, Gamma: 0.9})
	s := obsState{0.5}
	// A terminal reward of 1 for action 1 raises its value by alpha*(1-0).
	a.Observe(s, vecAction(1), 1, s, true)
	if got := a.values(a.key(s))[1]; got != 0.5 {
		t.Fatalf("Q after update = %g, want 0.5", got)
	}
}

var _ core.Agent = New(DefaultConfig())
