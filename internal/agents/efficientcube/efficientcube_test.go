package efficientcube

import (
	"path/filepath"
	"slices"
	"testing"

	rubix "github.com/danielriddell21/rubix/pkg/cube"
)

func TestEncodeOneHot(t *testing.T) {
	x := encode(rubix.Solved())
	if len(x) != inputDim {
		t.Fatalf("encoding len = %d, want %d", len(x), inputDim)
	}
	ones := 0.0
	for _, v := range x {
		ones += v
	}
	if ones != 54 {
		t.Fatalf("one-hot has %g ones, want 54", ones)
	}
	for i := range 54 {
		var sum float64
		for k := range 6 {
			sum += x[i*6+k]
		}
		if sum != 1 {
			t.Fatalf("facelet %d not one-hot (sum %g)", i, sum)
		}
	}
}

func TestInverseRoundTrip(t *testing.T) {
	c := rubix.ScrambledCube(5, 1)
	for m := range numMoves {
		mv := rubix.Move(m)
		if c.Applied(mv).Applied(mv.Inverse()) != c {
			t.Fatalf("move %v then inverse did not restore the cube", mv)
		}
	}
}

func TestSolveSolvedIsNoop(t *testing.T) {
	p := NewPolicy([]int{32}, 1)
	r := p.Solve(rubix.Solved(), DefaultBeamConfig())
	if !r.Solved || len(r.Moves) != 0 {
		t.Fatalf("solving a solved cube: %+v", r)
	}
}

func TestLearnsAndSolves(t *testing.T) {
	p := Train(TrainConfig{Hidden: []int{128, 64}, K: 5, Batch: 256, Iters: 1000, LR: 1e-3, Seed: 1}, nil)

	solved := 0
	for i := range 40 {
		c := rubix.ScrambledCube(5, int64(1000+i))
		if p.Solve(c, BeamConfig{Width: 200, MaxDepth: 14}).Solved {
			solved++
		}
	}
	if solved < 34 { // ≥85% of held-out depth-5 scrambles
		t.Fatalf("beam search solved only %d/40 depth-5 scrambles", solved)
	}

	// Drive the beam agent through an episode: it plans from the first
	// observation, then replays its solution, exercising the observation→cube
	// reconstruction and action contract.
	agent := NewAgent(p, BeamConfig{Width: 200, MaxDepth: 14})
	c := rubix.ScrambledCube(3, 12345)
	for range 20 {
		if c.IsSolved() {
			break
		}
		a := agent.Act(stubState(faceletObs(c)))
		c = c.Applied(rubix.Move(int(a.Vector()[0] + 0.5)))
	}
	if !c.IsSolved() {
		t.Fatal("beam agent failed to solve a depth-3 scramble")
	}
}

func TestDeterminism(t *testing.T) {
	cfg := TrainConfig{Hidden: []int{64}, K: 4, Batch: 128, Iters: 150, LR: 1e-3, Seed: 7}
	a := Train(cfg, nil)
	b := Train(cfg, nil)
	c := rubix.ScrambledCube(4, 99)
	if !slices.Equal(a.Probs(c), b.Probs(c)) {
		t.Fatal("training is not deterministic for the same config and seed")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	p := Train(TrainConfig{Hidden: []int{64}, K: 4, Batch: 128, Iters: 120, LR: 1e-3, Seed: 3}, nil)
	path := filepath.Join(t.TempDir(), "policy.json")
	if err := p.Save(path); err != nil {
		t.Fatal(err)
	}
	q, err := LoadPolicy(path)
	if err != nil {
		t.Fatal(err)
	}
	c := rubix.ScrambledCube(4, 5)
	if !slices.Equal(p.Probs(c), q.Probs(c)) {
		t.Fatal("loaded policy differs from saved")
	}
}

type stubState []float64

func (s stubState) Observation() []float64 { return s }

func faceletObs(c rubix.Cube) []float64 {
	f := c.ToFacelets()
	o := make([]float64, len(f))
	for i, col := range f {
		o[i] = float64(col) / 5
	}
	return o
}
