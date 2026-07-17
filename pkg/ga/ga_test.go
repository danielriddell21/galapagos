package ga

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func TestRandomGenomeDeterministic(t *testing.T) {
	a := RandomGenome(10, rand.New(rand.NewPCG(1, 2)))
	b := RandomGenome(10, rand.New(rand.NewPCG(1, 2)))
	if len(a) != 10 {
		t.Fatalf("length = %d, want 10", len(a))
	}
	if !slices.Equal(a, b) {
		t.Fatal("same seed produced different genomes")
	}
}

func TestReproduceShapeAndElitism(t *testing.T) {
	cfg := Config{PopulationSize: 20, EliteFraction: 0.1, MutationRate: 0.05, MutationStd: 0.2, Seed: 42}
	genomes, fitness := seedPop(20, 5)
	fitness[7] = 1000 // make member 7 the unambiguous champion

	next := Reproduce(cfg, genomes, fitness, 0, 0)
	if len(next) != 20 {
		t.Fatalf("next generation size = %d, want 20", len(next))
	}
	if !containsVec(next, genomes[7]) {
		t.Fatal("fittest genome was not carried through as an elite")
	}
}

func TestReproduceDeterministic(t *testing.T) {
	cfg := Config{PopulationSize: 16, EliteFraction: 0.1, MutationRate: 0.1, MutationStd: 0.3, Seed: 7}
	genomes, fitness := seedPop(16, 6)
	a := Reproduce(cfg, genomes, fitness, 2, 1)
	b := Reproduce(cfg, genomes, fitness, 2, 1)
	for i := range a {
		if !slices.Equal(a[i], b[i]) {
			t.Fatalf("Reproduce not deterministic at genome %d", i)
		}
	}
}

func TestReproduceInjectsImmigrants(t *testing.T) {
	cfg := Config{PopulationSize: 20, EliteFraction: 0.1, ImmigrantFraction: 0.15, MutationRate: 0.05, MutationStd: 0.2, Seed: 42}
	const n, glen = 20, 5
	genomes, fitness := seedPop(n, glen)
	next := Reproduce(cfg, genomes, fitness, 0, 0)

	// The immigrant slots (just after the elites) must be the fresh random genomes
	// drawn from the dedicated immigrant stream for generation 0.
	eliteCount := max(1, int(cfg.EliteFraction*float64(n)))
	immigrantCount := int(cfg.ImmigrantFraction * float64(n))
	if immigrantCount == 0 {
		t.Fatal("expected at least one immigrant")
	}
	for i := range immigrantCount {
		irng := rand.New(rand.NewPCG(uint64(cfg.Seed)^streamImmigrant, uint64(i)))
		want := RandomGenome(glen, irng)
		if !slices.Equal(next[eliteCount+i], want) {
			t.Fatalf("immigrant %d is not the expected fresh random genome", i)
		}
	}
}

func TestAdaptiveMutationWidensOnStagnation(t *testing.T) {
	cfg := Config{MutationRate: 0.05, MutationStd: 0.2, StagnationWindow: 4, HyperMutation: 6}
	if r, s := adaptiveMutation(cfg, 0); r != 0.05 || s != 0.2 {
		t.Fatalf("fresh: got %v/%v, want base rates", r, s)
	}
	r, s := adaptiveMutation(cfg, 4)
	wantR := min(1, cfg.MutationRate*cfg.HyperMutation)
	wantS := cfg.MutationStd * cfg.HyperMutation
	if r != wantR || s != wantS {
		t.Fatalf("stalled: got %v/%v, want %v/%v", r, s, wantR, wantS)
	}
}

func TestRunMaximizesAndNeverBelowIncumbent(t *testing.T) {
	// Maximize -sum((g-target)^2): a smooth concave objective peaking at target.
	target := []float64{0.5, -1.0, 2.0}
	f := func(g []float64) float64 {
		var s float64
		for i := range g {
			d := g[i] - target[i]
			s += d * d
		}
		return -s
	}
	cfg := Config{PopulationSize: 40, EliteFraction: 0.1, MutationRate: 0.2, MutationStd: 0.3, Seed: 1}
	incumbent := []float64{0, 0, 0}
	res := Run(cfg, len(target), 80, f, incumbent)

	if res.Fitness < f(incumbent) {
		t.Fatalf("result fitness %.4f is worse than the incumbent %.4f", res.Fitness, f(incumbent))
	}
	if res.Fitness < -0.05 {
		t.Fatalf("did not converge near the optimum: fitness %.4f", res.Fitness)
	}
}

func TestRunDeterministic(t *testing.T) {
	f := func(g []float64) float64 { return -g[0] * g[0] }
	cfg := Config{PopulationSize: 20, MutationRate: 0.1, MutationStd: 0.3, Seed: 5}
	a := Run(cfg, 1, 20, f)
	b := Run(cfg, 1, 20, f)
	if !slices.Equal(a.Best, b.Best) || a.Fitness != b.Fitness {
		t.Fatal("Run is not deterministic for a fixed seed")
	}
}

func seedPop(n, glen int) ([][]float64, []float64) {
	rng := rand.New(rand.NewPCG(9, 9))
	g := make([][]float64, n)
	f := make([]float64, n)
	for i := range g {
		g[i] = RandomGenome(glen, rng)
		f[i] = rng.Float64()
	}
	return g, f
}

func containsVec(pop [][]float64, v []float64) bool {
	for _, p := range pop {
		if slices.Equal(p, v) {
			return true
		}
	}
	return false
}
