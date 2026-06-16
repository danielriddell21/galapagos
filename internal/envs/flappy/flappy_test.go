package flappy

import (
	"math/rand/v2"
	"testing"

	"github.com/danielriddell21/galapagos/internal/render"
)

func newRNG() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

// act is a one-element action vector selecting a discrete action.
type act float64

func (a act) Vector() []float64 { return []float64{float64(a)} }

func TestResetDeterministicCourse(t *testing.T) {
	a := New(DefaultConfig())
	a.Reset(newRNG())
	b := New(DefaultConfig())
	b.Reset(newRNG())
	for i := range a.pipes {
		if a.pipes[i] != b.pipes[i] {
			t.Fatalf("same seed produced different pipe %d: %+v vs %+v", i, a.pipes[i], b.pipes[i])
		}
	}
}

func TestGravityPullsDownFlapLifts(t *testing.T) {
	e := New(DefaultConfig())
	e.Reset(newRNG())
	y0 := e.birdY
	e.Step(act(0)) // idle: gravity increases y (falls)
	if e.birdY <= y0 {
		t.Fatalf("gravity did not lower the bird: %v -> %v", y0, e.birdY)
	}
	// A flap sets an upward velocity, so the next step rises relative to falling.
	e.Step(act(1))
	if e.birdVY >= 0 {
		t.Fatalf("flap did not produce upward velocity: vy=%v", e.birdVY)
	}
}

func TestCollisionEndsEpisode(t *testing.T) {
	e := New(DefaultConfig())
	e.Reset(newRNG())
	// Idle until the bird hits the ground (or a pipe); it must terminate.
	done := false
	for range DefaultConfig().MaxSteps {
		_, _, d := e.Step(act(0))
		if d {
			done = true
			break
		}
	}
	if !done {
		t.Fatal("falling bird never ended the episode")
	}
	if !e.crashed() && e.t < DefaultConfig().MaxSteps {
		t.Fatal("episode ended without a crash before the step budget")
	}
}

func TestSurvivingStepRewarded(t *testing.T) {
	e := New(DefaultConfig())
	e.Reset(newRNG())
	_, r, done := e.Step(act(1)) // one flap keeps the bird aloft for this step
	if done {
		t.Fatal("a single step should not crash from the center")
	}
	if r < stepReward {
		t.Fatalf("survival reward = %v, want >= %v", float64(r), stepReward)
	}
}

func TestActionSpecDiscrete(t *testing.T) {
	spec := New(DefaultConfig()).ActionSpec()
	if !spec.Discrete || spec.Dim != 1 {
		t.Fatalf("unexpected action spec %+v", spec)
	}
	if got := int(spec.High[0]-spec.Low[0]) + 1; got != 2 {
		t.Fatalf("discrete action count = %d, want 2", got)
	}
}

func TestObservationWithinBounds(t *testing.T) {
	e := New(DefaultConfig())
	s := e.Reset(newRNG())
	obs := s.Observation()
	if len(obs) != 4 {
		t.Fatalf("observation len = %d, want 4", len(obs))
	}
	for _, v := range obs {
		if v < 0 || v > 1 {
			t.Fatalf("observation %g out of [0,1]", v)
		}
	}
}

func TestClearingPipeCountsAndScores(t *testing.T) {
	e := New(DefaultConfig())
	e.Reset(newRNG())
	if got := e.Status(); got != "pipes 0" {
		t.Fatalf("fresh episode status = %q, want %q", got, "pipes 0")
	}

	// Place a single pipe one step away from passing the bird's x, with the bird
	// centered in its gap so it clears cleanly rather than crashing.
	e.birdY, e.birdVY = worldH/2, 0
	e.pipes = []pipe{{x: birdX - pipeWidth + pipeSpeed - 0.01, gapTop: worldH/2 - gapHeight/2}}

	_, r, done := e.Step(act(0))
	if done {
		t.Fatal("a bird inside the gap should not crash while clearing a pipe")
	}
	if e.cleared != 1 {
		t.Fatalf("cleared = %d, want 1", e.cleared)
	}
	if float64(r) < clearBonus {
		t.Fatalf("clearing reward = %v, want >= %v", float64(r), clearBonus)
	}
	if got := e.Status(); got != "pipes 1" {
		t.Fatalf("status after clearing = %q, want %q", got, "pipes 1")
	}
}

func TestRenderDoesNotPanic(t *testing.T) {
	e := New(DefaultConfig())
	e.Reset(newRNG())
	e.Render(render.NewNop()) // exercises the pipe + bird draw path
}
