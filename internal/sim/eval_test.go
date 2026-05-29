package sim

import (
	"runtime"
	"slices"
	"testing"

	"github.com/danielriddell21/galapagos/internal/core"
)

const evalSeed = 42

// factory returns an EnvFactory producing independent fake environments.
func factory() EnvFactory {
	return func() core.MultiEnvironment { return newFakeMultiEnv() }
}

// membersOf collects a fake population's members as core.Individual values.
func membersOf(p *fakePop) []core.Individual {
	out := make([]core.Individual, len(p.members))
	for i, m := range p.members {
		out[i] = m
	}
	return out
}

func TestRunGenerationFitness(t *testing.T) {
	pop := newFakePop(3, 5, 1, 8)
	got := RunGeneration(newFakeMultiEnv(), pop, 100, evalSeed, nil)
	// Each member lives for genome[0] steps, earning 1 reward per step.
	if want := []float64{3, 5, 1, 8}; !slices.Equal(got, want) {
		t.Fatalf("fitness = %v, want %v", got, want)
	}
	for i, m := range pop.members {
		if float64(m.Fitness()) != got[i] {
			t.Fatalf("member %d fitness not recorded: %v", i, m.Fitness())
		}
	}
}

func TestSimultaneousMatchesParallel(t *testing.T) {
	lifespans := []float64{3, 5, 1, 8, 2, 7}
	sequential := RunGeneration(newFakeMultiEnv(), newFakePop(lifespans...), 100, evalSeed, nil)
	parallel := EvaluateParallel(factory(), membersOf(newFakePop(lifespans...)), 100, evalSeed)
	if !slices.Equal(sequential, parallel) {
		t.Fatalf("parallel %v != simultaneous %v", parallel, sequential)
	}
}

func TestParallelDeterministicAcrossWorkers(t *testing.T) {
	lifespans := []float64{3, 5, 1, 8, 2, 7, 4, 6}
	old := runtime.GOMAXPROCS(1)
	one := EvaluateParallel(factory(), membersOf(newFakePop(lifespans...)), 100, evalSeed)
	runtime.GOMAXPROCS(8)
	many := EvaluateParallel(factory(), membersOf(newFakePop(lifespans...)), 100, evalSeed)
	runtime.GOMAXPROCS(old)
	if !slices.Equal(one, many) {
		t.Fatalf("GOMAXPROCS changed result: 1=>%v 8=>%v", one, many)
	}
}
