package nn

// Grads holds gradients with the same shape as an MLP's weights and biases.
type Grads struct {
	dW [][]float64
	dB [][]float64
}

// newGrads allocates zeroed gradients matching m.
func (m *MLP) newGrads() *Grads {
	g := &Grads{dW: make([][]float64, len(m.layers)), dB: make([][]float64, len(m.layers))}
	for l, ly := range m.layers {
		g.dW[l] = make([]float64, len(ly.w))
		g.dB[l] = make([]float64, len(ly.b))
	}
	return g
}

// backwardInto accumulates the gradients of a single example into g, given the
// cached forward pass and the loss gradient w.r.t. the output activations.
func (m *MLP) backwardInto(g *Grads, acts, zs [][]float64, dOut []float64) {
	delta := dOut
	for l := len(m.layers) - 1; l >= 0; l-- {
		ly := m.layers[l]
		dPre := make([]float64, ly.out)
		for j := range ly.out {
			dPre[j] = delta[j] * deriv(zs[l][j], ly.act)
		}
		aPrev := acts[l]
		dW, dB := g.dW[l], g.dB[l]
		for i := range ly.in {
			a := aPrev[i]
			row := i * ly.out
			for j := range ly.out {
				dW[row+j] += a * dPre[j]
			}
		}
		for j := range ly.out {
			dB[j] += dPre[j]
		}
		if l > 0 {
			dPrev := make([]float64, ly.in)
			for i := range ly.in {
				row := i * ly.out
				s := 0.0
				for j := range ly.out {
					s += ly.w[row+j] * dPre[j]
				}
				dPrev[i] = s
			}
			delta = dPrev
		}
	}
}

// scale divides all gradients by n (to average over a minibatch).
func (g *Grads) scale(n float64) {
	for l := range g.dW {
		for i := range g.dW[l] {
			g.dW[l][i] /= n
		}
		for j := range g.dB[l] {
			g.dB[l][j] /= n
		}
	}
}
