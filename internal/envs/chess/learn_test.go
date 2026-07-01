package chess_test

import (
	"testing"

	"github.com/danielriddell21/galapagos/internal/agents/ga"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/chess"
	"github.com/danielriddell21/galapagos/internal/sim"
)

func TestGAEvolvesAgainstRandom(t *testing.T) {
	pop := ga.New(ga.Config{
		Population: 24, EliteFraction: 0.1, MutationRate: 0.05, MutationStd: 0.2,
		HiddenSize: 8, Inputs: 64, Outputs: 128, Seed: 7,
	})
	const maxPlies = 40
	factory := sim.EnvFactory(func() core.MultiEnvironment {
		return sim.AsMulti(func() core.Environment { return chess.New(chess.Config{MaxPlies: maxPlies}) })
	})
	members := func() []core.Individual {
		var out []core.Individual
		for m := range pop.All() {
			out = append(out, m)
		}
		return out
	}

	best := func(fs []float64) float64 {
		m := fs[0]
		for _, f := range fs {
			if f > m {
				m = f
			}
		}
		return m
	}

	peak := best(sim.EvaluateParallel(factory, members(), maxPlies, 7))
	for range 8 {
		pop.Evolve()
		peak = max(peak, best(sim.EvaluateParallel(factory, members(), maxPlies, 7)))
	}
	if peak < 0.5 {
		t.Fatalf("GA did not learn to win material vs random: peak best fitness %.3f", peak)
	}
}
