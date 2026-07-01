package nn

import "math"

type Adam struct {
	lr, b1, b2, eps float64
	t               int
	mW, vW          [][]float64
	mB, vB          [][]float64
}

func NewAdam(m *MLP, lr float64) *Adam {
	a := &Adam{lr: lr, b1: 0.9, b2: 0.999, eps: 1e-8}
	for _, ly := range m.layers {
		a.mW = append(a.mW, make([]float64, len(ly.w)))
		a.vW = append(a.vW, make([]float64, len(ly.w)))
		a.mB = append(a.mB, make([]float64, len(ly.b)))
		a.vB = append(a.vB, make([]float64, len(ly.b)))
	}
	return a
}

func (a *Adam) Step(m *MLP, g *Grads) {
	a.t++
	bc1 := 1 - math.Pow(a.b1, float64(a.t))
	bc2 := 1 - math.Pow(a.b2, float64(a.t))
	for l := range m.layers {
		adam(m.layers[l].w, g.dW[l], a.mW[l], a.vW[l], a, bc1, bc2)
		adam(m.layers[l].b, g.dB[l], a.mB[l], a.vB[l], a, bc1, bc2)
	}
}

func adam(p, grad, mom, vel []float64, a *Adam, bc1, bc2 float64) {
	for i := range p {
		mom[i] = a.b1*mom[i] + (1-a.b1)*grad[i]
		vel[i] = a.b2*vel[i] + (1-a.b2)*grad[i]*grad[i]
		mHat := mom[i] / bc1
		vHat := vel[i] / bc2
		p[i] -= a.lr * mHat / (math.Sqrt(vHat) + a.eps)
	}
}
