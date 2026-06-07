package efficientcube

import (
	"runtime"
	"slices"
	"testing"

	rubix "github.com/danielriddell21/rubix/pkg/cube"
)

func batchCubes(n int) []rubix.Cube {
	cs := make([]rubix.Cube, n)
	for i := range n {
		cs[i] = rubix.ScrambledCube(4, int64(500+i))
	}
	return cs
}

// TestSolveBatchMatchesSequential checks the parallel batch solve produces the
// same result as solving each cube sequentially. An untrained policy suffices:
// the property must hold regardless of policy quality (and keeps the test fast
// while still exercising concurrent Forward calls under -race).
func TestSolveBatchMatchesSequential(t *testing.T) {
	p := NewPolicy([]int{96}, 2)
	cfg := BeamConfig{Width: 100, MaxDepth: 8}
	cubes := batchCubes(16)

	got := p.SolveBatch(cubes, cfg)
	for i, c := range cubes {
		want := p.Solve(c, cfg)
		if got[i].Solved != want.Solved || !slices.Equal(got[i].Moves, want.Moves) {
			t.Fatalf("cube %d: batch %+v != sequential %+v", i, got[i], want)
		}
	}
}

func TestSolveBatchWorkerIndependent(t *testing.T) {
	p := NewPolicy([]int{96}, 2)
	cfg := BeamConfig{Width: 100, MaxDepth: 8}
	cubes := batchCubes(16)

	old := runtime.GOMAXPROCS(1)
	one := p.SolveBatch(cubes, cfg)
	runtime.GOMAXPROCS(4)
	many := p.SolveBatch(cubes, cfg)
	runtime.GOMAXPROCS(old)

	for i := range cubes {
		if one[i].Solved != many[i].Solved || !slices.Equal(one[i].Moves, many[i].Moves) {
			t.Fatalf("cube %d differs across worker counts", i)
		}
	}
}
