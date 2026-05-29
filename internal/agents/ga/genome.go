package ga

import (
	"math/rand/v2"
	"slices"
)

// randomGenome returns a genome of length n with Gaussian-initialized weights.
func randomGenome(n int, rng *rand.Rand) []float64 {
	g := make([]float64, n)
	for i := range n {
		g[i] = rng.NormFloat64()
	}
	return g
}

// crossover blends two parent genomes. With uniform probability per gene it
// either averages the parents (blend) or copies one parent's gene, producing a
// child that mixes both. Parents must be the same length.
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

// mutate perturbs genes in place with Gaussian noise of the given standard
// deviation, applying the perturbation to each gene with probability rate.
func mutate(g []float64, rate, std float64, rng *rand.Rand) {
	for i := range g {
		if rng.Float64() < rate {
			g[i] += rng.NormFloat64() * std
		}
	}
}

// clone returns an independent copy of a genome so elites survive unmutated.
func clone(g []float64) []float64 { return slices.Clone(g) }
