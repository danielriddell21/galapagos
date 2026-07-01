package cube

import (
	"math/rand/v2"
	"testing"

	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/render"
)

func newRNG() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

type move float64

func (m move) Vector() []float64 { return []float64{float64(m)} }

func TestResetDeterministicAndScrambled(t *testing.T) {
	a := New(DefaultConfig()).Reset(newRNG()).(State)
	b := New(DefaultConfig()).Reset(newRNG()).(State)
	if a != b {
		t.Fatal("same seed must produce the same scramble")
	}
	if a.c.IsSolved() {
		t.Fatal("a non-zero scramble must not start solved")
	}
}

func TestStateIsComparable(t *testing.T) {
	// The state must be usable as a map key (the whole point of the cube env).
	seen := map[core.State]int{}
	e := New(DefaultConfig())
	s := e.Reset(newRNG())
	seen[s]++
	seen[s]++
	if seen[s] != 2 {
		t.Fatalf("state not stable as a map key: %d", seen[s])
	}
}

func TestStepSolvesWhenInverted(t *testing.T) {
	// Scramble one move, then applying its inverse must solve and reward +1.
	e := New(Config{ScrambleDepth: 1, MaxSteps: 5})
	start := e.Reset(newRNG()).(State)
	// Find the single move that solves it by trying all inverses.
	solvedReward := core.Reward(0)
	done := false
	for m := range int(18) {
		probe := New(Config{ScrambleDepth: 1, MaxSteps: 5})
		probe.Reset(newRNG()) // identical scramble (same seed)
		_, r, d := probe.Step(move(m))
		if d && probe.Solved() {
			solvedReward, done = r, true
			break
		}
	}
	if !done {
		t.Fatal("no single move solved a depth-1 scramble")
	}
	if solvedReward != 1 {
		t.Fatalf("solve reward = %v, want 1", solvedReward)
	}
	_ = start
}

func TestEpisodeEndsAtMaxSteps(t *testing.T) {
	e := New(Config{ScrambleDepth: 6, MaxSteps: 3})
	e.Reset(newRNG())
	var done bool
	for range 3 {
		_, _, done = e.Step(move(0)) // U repeatedly is unlikely to solve
	}
	if !done {
		t.Fatal("episode should end at the step budget")
	}
}

func TestObservationWithinBounds(t *testing.T) {
	s := New(DefaultConfig()).Reset(newRNG())
	obs := s.Observation()
	if len(obs) != 54 {
		t.Fatalf("observation len = %d, want 54", len(obs))
	}
	for _, v := range obs {
		if v < 0 || v > 1 {
			t.Fatalf("observation %g out of [0,1]", v)
		}
	}
}

func TestActionSpecDiscrete(t *testing.T) {
	spec := New(DefaultConfig()).ActionSpec()
	if !spec.Discrete || spec.Dim != 1 {
		t.Fatalf("unexpected action spec %+v", spec)
	}
	if got := int(spec.High[0]-spec.Low[0]) + 1; got != 18 {
		t.Fatalf("discrete action count = %d, want 18", got)
	}
}

func TestRenderDoesNotPanic(t *testing.T) {
	e := New(DefaultConfig())
	e.Reset(newRNG())
	e.Render(render.NewNop()) // exercises ToFacelets + the draw path
}
