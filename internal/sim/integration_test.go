package sim_test

import (
	"runtime"
	"slices"
	"testing"

	"github.com/danielriddell21/galapagos/internal/agents/ga"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/racing"
	"github.com/danielriddell21/galapagos/internal/sim"
)

const (
	itSeed     = 42
	itMaxSteps = 400
	itPop      = 16
)

func gaConfig() ga.Config {
	rays := racing.DefaultSensorParams().Count
	return ga.Config{
		Population:    itPop,
		EliteFraction: 0.1,
		MutationRate:  0.05,
		MutationStd:   0.2,
		HiddenSize:    8,
		Inputs:        rays + 1,
		Outputs:       2,
		Seed:          itSeed,
	}
}

func racingFactory() sim.EnvFactory {
	return func() core.MultiEnvironment { return racing.New(racing.DefaultConfig()) }
}

func members(p *ga.Population) []core.Individual {
	var out []core.Individual
	for m := range p.All() {
		out = append(out, m)
	}
	return out
}

// TestSimultaneousEqualsParallelOnRacing locks the property the determinism
// contract depends on: cars only collide with walls, so evaluating the whole
// population together on one track yields exactly the same fitness as rolling
// each car out alone.
func TestSimultaneousEqualsParallelOnRacing(t *testing.T) {
	sequential := sim.RunGeneration(racing.New(racing.DefaultConfig()), ga.New(gaConfig()), itMaxSteps, itSeed, nil)
	parallel := sim.EvaluateParallel(racingFactory(), members(ga.New(gaConfig())), itMaxSteps, itSeed)
	if !slices.Equal(sequential, parallel) {
		t.Fatalf("simultaneous and parallel fitness differ:\n seq=%v\n par=%v", sequential, parallel)
	}
}

// TestParallelDeterministicAcrossWorkers verifies fitness is independent of the
// worker count.
func TestParallelDeterministicAcrossWorkers(t *testing.T) {
	old := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(old)
	one := sim.EvaluateParallel(racingFactory(), members(ga.New(gaConfig())), itMaxSteps, itSeed)
	runtime.GOMAXPROCS(8)
	many := sim.EvaluateParallel(racingFactory(), members(ga.New(gaConfig())), itMaxSteps, itSeed)
	if !slices.Equal(one, many) {
		t.Fatalf("worker count changed fitness:\n  1=%v\n  8=%v", one, many)
	}
}

// TestHeadlessTrainingReproducible trains the GA end to end twice and requires
// the evolved best genome to be byte-identical, regardless of GOMAXPROCS.
func TestHeadlessTrainingReproducible(t *testing.T) {
	train := func(workers int) []float64 {
		old := runtime.GOMAXPROCS(workers)
		defer runtime.GOMAXPROCS(old)
		pop := ga.New(gaConfig())
		sim.TrainHeadless(racingFactory(), pop, sim.RunConfig{Seed: itSeed, MaxSteps: itMaxSteps, Generations: 5}, nil)
		return pop.Best()
	}
	if a, b := train(1), train(4); !slices.Equal(a, b) {
		t.Fatal("headless training is not reproducible across worker counts")
	}
}
