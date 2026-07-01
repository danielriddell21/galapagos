package sim

import (
	"testing"
	"testing/synctest"
	"time"
)

func TestPaceTiming(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		count := 0
		Pace(5, 10*time.Millisecond, func() { count++ })
		if count != 5 {
			t.Fatalf("step count = %d, want 5", count)
		}
		if elapsed := time.Since(start); elapsed != 40*time.Millisecond {
			t.Fatalf("elapsed = %v, want 40ms", elapsed)
		}
	})
}

func TestTrainHeadlessAdvancesGenerations(t *testing.T) {
	pop := newFakePop(3, 5, 1)
	tel := NewTelemetry(16, nil)
	TrainHeadless(factory(), pop, RunConfig{Seed: evalSeed, MaxSteps: 100, Generations: 4}, tel)
	if pop.Generation() != 4 {
		t.Fatalf("generation = %d, want 4", pop.Generation())
	}
	if got := len(tel.History()); got != 4 {
		t.Fatalf("history length = %d, want 4", got)
	}
}
