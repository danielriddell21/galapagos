package neat_test

import (
	"slices"
	"testing"

	"github.com/danielriddell21/galapagos/internal/agents/neat"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/racing"
	"github.com/danielriddell21/galapagos/internal/sim"
)

func neatConfig() neat.Config {
	c := neat.DefaultConfig()
	c.Population = 40
	c.Inputs = racing.DefaultSensorParams().Count + 1
	c.Outputs = 2
	c.Seed = 9
	return c
}

func racingFactory() sim.EnvFactory {
	return func() core.MultiEnvironment { return racing.New(racing.DefaultConfig()) }
}

func members(p *neat.Population) []core.Individual {
	var out []core.Individual
	for m := range p.All() {
		out = append(out, m)
	}
	return out
}

// TestNEATImprovesAndIsReproducible evolves topologies on the racing task and
// requires the best fitness to improve, and two identically-seeded runs to
// produce identical best fitness.
func TestNEATImprovesAndIsReproducible(t *testing.T) {
	run := func() (before, after float64) {
		pop := neat.New(neatConfig())
		before = slices.Max(sim.EvaluateParallel(racingFactory(), members(pop), 400, 9))
		sim.TrainHeadless(racingFactory(), pop, sim.RunConfig{Seed: 9, MaxSteps: 400, Generations: 12}, nil)
		after = slices.Max(sim.EvaluateParallel(racingFactory(), members(pop), 400, 9))
		return before, after
	}
	b1, a1 := run()
	if a1 <= b1 {
		t.Fatalf("NEAT did not improve: before=%.3f after=%.3f", b1, a1)
	}
	b2, a2 := run()
	if b1 != b2 || a1 != a2 {
		t.Fatalf("NEAT not reproducible: run1=(%.6f,%.6f) run2=(%.6f,%.6f)", b1, a1, b2, a2)
	}
}
