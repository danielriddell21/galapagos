// Package ga is a small, dependency-free genetic algorithm for maximizing a
// caller-supplied fitness function over fixed-length vectors of float64.
//
// It combines elitism, tournament selection, per-gene crossover, and Gaussian
// mutation, and resists premature convergence with two diversity mechanisms: a
// floor of fresh random immigrants injected every generation, and a mutation
// rate that widens while the best fitness is stagnating and relaxes once it
// improves again. All randomness derives from [Config.Seed], so a run is fully
// reproducible and independent of the order in which genomes are evaluated.
//
// Most callers use [Run] (or an [Optimizer] for step-by-step control). Callers
// that need to evaluate a generation concurrently can drive [Reproduce], the
// pure generational operator, from their own loop.
package ga

import (
	"math/rand/v2"
	"slices"
)

// Distinct PCG stream constants keep genome initialization, breeding, and
// immigrant generation independent while all deriving from the seed.
const (
	streamInit      uint64 = 0x53c5ff8e3d9b1a77
	streamEvolve    uint64 = 0x6a09e667f3bcc908
	streamImmigrant uint64 = 0x9e3779b97f4a7c15
)

// Config parameters the genetic algorithm. Zero values for TournamentSize and
// the diversity fields are replaced with sensible defaults, so the minimum
// useful configuration is a PopulationSize, the two mutation settings, and a
// Seed.
type Config struct {
	// PopulationSize is the number of genomes carried each generation. It is used
	// by [Run] and [NewOptimizer]; [Reproduce] takes the size from its input.
	PopulationSize int
	// EliteFraction is the share of the fittest genomes copied unchanged into the
	// next generation. At least one elite is always kept.
	EliteFraction float64
	// MutationRate is the per-gene probability of a Gaussian perturbation.
	MutationRate float64
	// MutationStd is the standard deviation of that perturbation.
	MutationStd float64
	// TournamentSize is the number of genomes sampled per selection (default 3).
	TournamentSize int
	// ImmigrantFraction is the share of each generation reseeded with fresh random
	// genomes, keeping a permanent floor of diversity (default 0.15).
	ImmigrantFraction float64
	// StagnationWindow is the number of generations without an improvement in the
	// best fitness that trigger widened mutation (default 4; 0 disables it).
	StagnationWindow int
	// HyperMutation scales MutationRate and MutationStd while the best fitness is
	// stagnating (default 6).
	HyperMutation float64
	// Seed makes the whole run reproducible.
	Seed int64
}

func (c Config) withDefaults() Config {
	if c.TournamentSize <= 0 {
		c.TournamentSize = 3
	}
	if c.ImmigrantFraction == 0 {
		c.ImmigrantFraction = 0.15
	}
	if c.StagnationWindow == 0 {
		c.StagnationWindow = 4
	}
	if c.HyperMutation == 0 {
		c.HyperMutation = 6
	}
	return c
}

// Fitness scores a genome; larger is better.
type Fitness func(genome []float64) float64

// RandomGenome returns n independent standard-normal genes drawn from rng.
func RandomGenome(n int, rng *rand.Rand) []float64 {
	g := make([]float64, n)
	for i := range n {
		g[i] = rng.NormFloat64()
	}
	return g
}

// Reproduce returns the next generation from the current genomes and their
// fitness (in the same order). gen is the generation index, which seeds the
// immigrant stream; stalled is the number of generations since the best fitness
// last improved, which drives adaptive mutation. It is pure and deterministic
// given Config.Seed, gen, stalled, and the inputs, so a caller that evaluates
// fitness concurrently still gets reproducible evolution.
func Reproduce(cfg Config, genomes [][]float64, fitness []float64, gen, stalled int) [][]float64 {
	cfg = cfg.withDefaults()
	n := len(genomes)
	rng := rand.New(rand.NewPCG(uint64(cfg.Seed)^streamEvolve, uint64(gen)))
	order := rankDescending(fitness)
	rate, std := adaptiveMutation(cfg, stalled)

	eliteCount := max(1, int(cfg.EliteFraction*float64(n)))
	immigrantCount := int(cfg.ImmigrantFraction * float64(n))
	if eliteCount+immigrantCount > n {
		immigrantCount = n - eliteCount
	}

	genomeLen := len(genomes[0])
	next := make([][]float64, 0, n)
	for i := range eliteCount {
		next = append(next, slices.Clone(genomes[order[i]]))
	}
	for i := range immigrantCount {
		irng := rand.New(rand.NewPCG(uint64(cfg.Seed)^streamImmigrant, uint64(gen)*uint64(n)+uint64(i)))
		next = append(next, RandomGenome(genomeLen, irng))
	}
	ranked := make([][]float64, n)
	rankedFit := make([]float64, n)
	for i, idx := range order {
		ranked[i], rankedFit[i] = genomes[idx], fitness[idx]
	}
	for len(next) < n {
		a := tournament(ranked, rankedFit, cfg.TournamentSize, rng)
		b := tournament(ranked, rankedFit, cfg.TournamentSize, rng)
		child := crossover(a, b, rng)
		mutate(child, rate, std, rng)
		next = append(next, child)
	}
	return next
}

// adaptiveMutation widens the breeding mutation while the champion is stuck.
func adaptiveMutation(cfg Config, stalled int) (rate, std float64) {
	rate, std = cfg.MutationRate, cfg.MutationStd
	if cfg.StagnationWindow > 0 && stalled >= cfg.StagnationWindow {
		rate = min(1, rate*cfg.HyperMutation)
		std *= cfg.HyperMutation
	}
	return rate, std
}

// rankDescending returns genome indices ordered by fitness, fittest first.
func rankDescending(fitness []float64) []int {
	order := make([]int, len(fitness))
	for i := range order {
		order[i] = i
	}
	slices.SortFunc(order, func(i, j int) int {
		switch {
		case fitness[j] < fitness[i]:
			return -1
		case fitness[j] > fitness[i]:
			return 1
		default:
			return 0
		}
	})
	return order
}

// tournament samples size genomes and returns the fittest, biasing selection
// toward stronger genomes while preserving diversity.
func tournament(genomes [][]float64, fitness []float64, size int, rng *rand.Rand) []float64 {
	best := rng.IntN(len(genomes))
	for range size - 1 {
		c := rng.IntN(len(genomes))
		if fitness[c] > fitness[best] {
			best = c
		}
	}
	return genomes[best]
}

// crossover blends two parents per gene: parent a, parent b, or their average.
func crossover(a, b []float64, rng *rand.Rand) []float64 {
	child := make([]float64, len(a))
	for i := range a {
		switch rng.IntN(3) {
		case 0:
			child[i] = a[i]
		case 1:
			child[i] = b[i]
		default:
			child[i] = (a[i] + b[i]) / 2
		}
	}
	return child
}

// mutate perturbs each gene with probability rate by Gaussian noise of the given
// standard deviation.
func mutate(g []float64, rate, std float64, rng *rand.Rand) {
	for i := range g {
		if rng.Float64() < rate {
			g[i] += rng.NormFloat64() * std
		}
	}
}
