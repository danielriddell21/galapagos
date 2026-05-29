// Package ga implements a genetic algorithm: a population of fixed-topology
// feedforward neural networks whose weights are evolved. It depends only on
// core, so it can drive any environment without importing one.
package ga

import "math"

// net is a single-hidden-layer feedforward network. Its weights and biases are
// a flat genome slice, which makes crossover and mutation simple vector ops.
type net struct {
	nIn, nHidden, nOut int
	genome             []float64
}

// GenomeLen returns the number of weights for a network of the given shape:
// input→hidden and hidden→output dense layers, each with biases.
func GenomeLen(nIn, nHidden, nOut int) int {
	return nIn*nHidden + nHidden + nHidden*nOut + nOut
}

// newNet wraps a genome as a network of the given shape. The genome must have
// length GenomeLen(nIn, nHidden, nOut).
func newNet(nIn, nHidden, nOut int, genome []float64) *net {
	return &net{nIn: nIn, nHidden: nHidden, nOut: nOut, genome: genome}
}

// forward evaluates the network. The hidden layer uses tanh; outputs are
// returned raw so the caller can map them to action ranges.
func (n *net) forward(x []float64) []float64 {
	w1 := n.genome[:n.nIn*n.nHidden]
	b1 := n.genome[n.nIn*n.nHidden : n.nIn*n.nHidden+n.nHidden]
	rest := n.genome[n.nIn*n.nHidden+n.nHidden:]
	w2 := rest[:n.nHidden*n.nOut]
	b2 := rest[n.nHidden*n.nOut:]

	hidden := make([]float64, n.nHidden)
	for j := range n.nHidden {
		sum := b1[j]
		for i := range n.nIn {
			sum += w1[i*n.nHidden+j] * x[i]
		}
		hidden[j] = math.Tanh(sum)
	}

	out := make([]float64, n.nOut)
	for k := range n.nOut {
		sum := b2[k]
		for j := range n.nHidden {
			sum += w2[j*n.nOut+k] * hidden[j]
		}
		out[k] = sum
	}
	return out
}
