package sim

import (
	"slices"
	"testing"
)

func TestLiveMatchesRunGeneration(t *testing.T) {
	lifespans := []float64{3, 5, 1, 8, 2, 7}
	batch := RunGeneration(newFakeMultiEnv(), newFakePop(lifespans...), 100, evalSeed, nil)

	live := NewLive(newFakeMultiEnv(), newFakePop(lifespans...), 100, evalSeed)
	for !live.Step() {
	}
	if !slices.Equal(batch, live.Fitness()) {
		t.Fatalf("live %v != batch %v", live.Fitness(), batch)
	}
	// The longest-lived member must be reported as best.
	if got := live.BestIndex(); got != 3 {
		t.Fatalf("BestIndex = %d, want 3", got)
	}
}
