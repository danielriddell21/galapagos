package ga

import "math"

type net struct {
	nIn, nHidden, nOut int
	genome             []float64
}

func GenomeLen(nIn, nHidden, nOut int) int {
	return nIn*nHidden + nHidden + nHidden*nOut + nOut
}

func newNet(nIn, nHidden, nOut int, genome []float64) *net {
	return &net{nIn: nIn, nHidden: nHidden, nOut: nOut, genome: genome}
}

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
