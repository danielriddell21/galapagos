package ga

import (
	"math/rand/v2"
	"slices"
)

// Optimizer maximizes a [Fitness] over fixed-length float genomes, one
// generation at a time. It tracks the best genome seen and the stagnation used
// by adaptive mutation. It is deterministic given [Config.Seed].
type Optimizer struct {
	cfg     Config
	genomes [][]float64
	fitness []float64
	gen     int

	bestGenome []float64
	bestScore  float64
	stalled    int
	seen       bool
}

// NewOptimizer builds an optimizer with a random initial population of the given
// genome length. Any seed genomes are placed at the front of the population
// (truncated to the population size), so a known-good starting point — such as
// the incumbent solution — is always evaluated and, with elitism, never lost.
func NewOptimizer(cfg Config, genomeLen int, seed ...[]float64) *Optimizer {
	cfg = cfg.withDefaults()
	size := max(cfg.PopulationSize, 1)
	genomes := make([][]float64, size)
	for i := range genomes {
		if i < len(seed) {
			genomes[i] = slices.Clone(seed[i])
			continue
		}
		rng := rand.New(rand.NewPCG(uint64(cfg.Seed)^streamInit, uint64(i)))
		genomes[i] = RandomGenome(genomeLen, rng)
	}
	return &Optimizer{cfg: cfg, genomes: genomes}
}

// Step evaluates the current population with f, updates the best genome and
// stagnation counter, then replaces the population with the next generation.
func (o *Optimizer) Step(f Fitness) {
	if o.fitness == nil {
		o.fitness = make([]float64, len(o.genomes))
	}
	best := 0
	for i, g := range o.genomes {
		o.fitness[i] = f(g)
		if o.fitness[i] > o.fitness[best] {
			best = i
		}
	}
	if top := o.fitness[best]; !o.seen || top > o.bestScore {
		o.bestScore, o.bestGenome, o.stalled, o.seen = top, slices.Clone(o.genomes[best]), 0, true
	} else {
		o.stalled++
	}
	o.genomes = Reproduce(o.cfg, o.genomes, o.fitness, o.gen, o.stalled)
	o.gen++
}

// Best returns a copy of the fittest genome found so far and its fitness. It
// returns nil before the first [Optimizer.Step].
func (o *Optimizer) Best() (genome []float64, fitness float64) {
	return slices.Clone(o.bestGenome), o.bestScore
}

// Generation returns the number of generations evaluated so far.
func (o *Optimizer) Generation() int { return o.gen }

// Result is the outcome of [Run].
type Result struct {
	// Best is the fittest genome found.
	Best []float64
	// Fitness is Best's fitness.
	Fitness float64
	// Generations is the number of generations evaluated.
	Generations int
}

// Run evolves a population for the given number of generations and returns the
// best genome found. Seed genomes are carried into the initial population (see
// [NewOptimizer]); passing the incumbent solution guarantees the result never
// scores worse than it, provided generations is at least one.
func Run(cfg Config, genomeLen, generations int, f Fitness, seed ...[]float64) Result {
	o := NewOptimizer(cfg, genomeLen, seed...)
	for range generations {
		o.Step(f)
	}
	genome, fitness := o.Best()
	return Result{Best: genome, Fitness: fitness, Generations: o.gen}
}
