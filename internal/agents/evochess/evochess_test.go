package evochess

import (
	"sync"
	"testing"

	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/chess"
	"github.com/danielriddell21/galapagos/internal/sim"
	gambit "github.com/danielriddell21/gambit/pkg/chess"
)

// boardState is a minimal chess state exposing the board to the search agent.
type boardState struct{ b *gambit.Board }

func (s boardState) Observation() []float64 { return nil }
func (s boardState) Board() *gambit.Board   { return s.b }

// seedGenome returns a genome with classical material and flat (zero)
// piece-square tables, so evaluation is pure material.
func seedGenome() []float64 {
	g := make([]float64, genomeLen)
	for t := range numTypes {
		g[t] = baseMaterial[t]
	}
	return g
}

func TestScoreFavorsMaterial(t *testing.T) {
	e := newEvaluator(seedGenome())
	b, err := gambit.ParseFEN("7k/8/8/8/8/8/8/Q6K w - - 0 1") // White up a queen
	if err != nil {
		t.Fatal(err)
	}
	if e.score(b) <= 0 {
		t.Fatalf("extra queen should score positive for White, got %g", e.score(b))
	}
}

func TestSearchWinsHangingPiece(t *testing.T) {
	e := newEvaluator(seedGenome())
	// White queen on a1, black rook hanging on a8 (king far away): Qxa8 wins it.
	b, err := gambit.ParseFEN("r6k/8/8/8/8/8/8/Q6K w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	m := bestMove(e, b, 2)
	a8, _ := gambit.ParseSquare("a8")
	if m.To() != a8 {
		t.Fatalf("search failed to win the hanging rook: chose %v", m)
	}
}

func TestActEncodesAMove(t *testing.T) {
	p := New(Config{Population: 4, Depth: 1, Seed: 1})
	a := p.Act(boardState{gambit.NewStartingBoard()})
	v := a.Vector()
	if len(v) != 128 {
		t.Fatalf("action length = %d, want 128", len(v))
	}
	var nonzero int
	for _, x := range v {
		if x != 0 {
			nonzero++
		}
	}
	if nonzero != 2 { // one "from" spike and one "to" spike
		t.Fatalf("encoded move has %d nonzero preferences, want 2", nonzero)
	}
}

func TestDeterministicFromSeed(t *testing.T) {
	move := func() gambit.Square {
		p := New(Config{Population: 4, Depth: 2, Seed: 5})
		return bestMove(p.best().eval, gambit.NewStartingBoard(), 2).To()
	}
	if move() != move() {
		t.Fatal("same seed produced different opening moves")
	}
}

func TestFrozenPolicyConcurrent(t *testing.T) {
	// Mirror the evaluator: each parallel rollout owns its own board, so the
	// shared frozen evaluator must be safe to read concurrently.
	p := New(Config{Population: 6, Depth: 2, Seed: 2})
	pol := p.FrozenPolicy() // snapshot of the best evaluator
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() { _ = pol(boardState{gambit.NewStartingBoard()}) })
	}
	wg.Wait()
}

// TestCoevolutionRaceFree runs the real parallel-evaluation path — every member
// searching against a frozen champion on its own board — under -race, to confirm
// the shared evaluator and per-rollout boards are safe.
func TestCoevolutionRaceFree(t *testing.T) {
	p := New(Config{Population: 8, Depth: 2, Seed: 3})
	champion := chess.Policy(p.FrozenPolicy())
	factory := sim.EnvFactory(func() core.MultiEnvironment {
		return sim.AsMulti(func() core.Environment {
			return chess.New(chess.Config{MaxPlies: 30, Opponent: champion})
		})
	})
	var members []core.Individual
	for m := range p.All() {
		members = append(members, m)
	}
	sim.EvaluateParallel(factory, members, 30, 3)
}

var _ core.PopulationAgent = New(DefaultConfig())
