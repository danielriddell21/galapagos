package cartpole

import (
	"math/rand/v2"
	"testing"

	"github.com/danielriddell21/galapagos/internal/core"
)

type act float64

func (a act) Vector() []float64 { return []float64{float64(a)} }

func newRNG() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

func TestResetObservationShape(t *testing.T) {
	e := New(DefaultConfig())
	s := e.Reset(newRNG())
	if got := len(s.Observation()); got != 4 {
		t.Fatalf("observation len = %d, want 4", got)
	}
	for _, v := range s.Observation() {
		if v < 0 || v > 1 {
			t.Fatalf("observation %g not normalized to [0,1]", v)
		}
	}
}

func TestPoleFallsWithConstantForce(t *testing.T) {
	// Always pushing the same direction must topple the pole before the budget.
	e := New(DefaultConfig())
	e.Reset(newRNG())
	done := false
	steps := 0
	for range DefaultConfig().MaxSteps {
		_, _, d := e.Step(act(1))
		steps++
		if d {
			done = true
			break
		}
	}
	if !done {
		t.Fatal("constant force should topple the pole")
	}
	if steps >= DefaultConfig().MaxSteps {
		t.Fatalf("episode ran the full budget (%d steps) without failing", steps)
	}
}

func TestDeterministicTrajectory(t *testing.T) {
	run := func() []float64 {
		e := New(DefaultConfig())
		e.Reset(newRNG())
		var trace []float64
		for range 50 {
			s, _, _ := e.Step(act(1))
			trace = append(trace, s.Observation()...)
		}
		return trace
	}
	a, b := run(), run()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("trajectory diverged at %d: %g vs %g", i, a[i], b[i])
		}
	}
}

var _ core.Environment = New(DefaultConfig())
