// Package nn is a small, deterministic, pure-Go neural network: a dense
// multi-layer perceptron with backpropagation, Adam, and softmax/cross-entropy.
// It exists because the genetic-algorithm net is forward-only; gradient-trained
// agents (such as the EfficientCube policy) need real training. It imports only
// the standard library.
package nn

import (
	"math"
	"math/rand/v2"
	"slices"
)

// Activation is a layer's nonlinearity.
type Activation uint8

const (
	// Linear applies no nonlinearity (used for logits/regression outputs).
	Linear Activation = iota
	// ReLU applies max(0, x).
	ReLU
)

// apply returns the activation of a pre-activation vector.
func apply(z []float64, a Activation) []float64 {
	out := make([]float64, len(z))
	switch a {
	case ReLU:
		for i, v := range z {
			out[i] = max(v, 0)
		}
	default:
		copy(out, z)
	}
	return out
}

// deriv returns d(activation)/d(preactivation) at z.
func deriv(z float64, a Activation) float64 {
	if a == ReLU {
		if z > 0 {
			return 1
		}
		return 0
	}
	return 1
}

// layer is a dense layer: weights w[i*out+j] (input i → output j), biases b[j].
type layer struct {
	in, out int
	w, b    []float64
	act     Activation
}

// MLP is a feedforward network of dense layers.
type MLP struct {
	layers []layer
	sizes  []int
	hidden Activation
	output Activation
}

// Config describes an MLP: layer sizes (input, hidden..., output), the hidden
// and output activations, and the seed for reproducible weight initialization.
type Config struct {
	Sizes  []int
	Hidden Activation
	Output Activation
	Seed   int64
}

// New builds an MLP with seeded He/Xavier initialization. Identical configs and
// seeds yield identical weights.
func New(cfg Config) *MLP {
	rng := rand.New(rand.NewPCG(uint64(cfg.Seed), 0x9e3779b97f4a7c15))
	m := &MLP{sizes: slices.Clone(cfg.Sizes), hidden: cfg.Hidden, output: cfg.Output}
	for l := range len(cfg.Sizes) - 1 {
		in, out := cfg.Sizes[l], cfg.Sizes[l+1]
		act := cfg.Hidden
		if l == len(cfg.Sizes)-2 {
			act = cfg.Output
		}
		// He for ReLU layers, Xavier otherwise.
		std := math.Sqrt(1.0 / float64(in))
		if act == ReLU {
			std = math.Sqrt(2.0 / float64(in))
		}
		w := make([]float64, in*out)
		for i := range w {
			w[i] = rng.NormFloat64() * std
		}
		m.layers = append(m.layers, layer{in: in, out: out, w: w, b: make([]float64, out), act: act})
	}
	return m
}

// InDim returns the input dimension.
func (m *MLP) InDim() int { return m.sizes[0] }

// OutDim returns the output dimension.
func (m *MLP) OutDim() int { return m.sizes[len(m.sizes)-1] }

// Forward evaluates the network and returns the output layer's values (raw
// logits when the output activation is Linear).
func (m *MLP) Forward(x []float64) []float64 {
	cur := x
	for _, ly := range m.layers {
		z := make([]float64, ly.out)
		for j := range ly.out {
			s := ly.b[j]
			for i := range ly.in {
				s += ly.w[i*ly.out+j] * cur[i]
			}
			z[j] = s
		}
		cur = apply(z, ly.act)
	}
	return cur
}

// forwardCached evaluates the network, retaining per-layer activations (acts,
// including the input) and pre-activations (zs) for backpropagation.
func (m *MLP) forwardCached(x []float64) (out []float64, acts, zs [][]float64) {
	acts = [][]float64{x}
	cur := x
	for _, ly := range m.layers {
		z := make([]float64, ly.out)
		for j := range ly.out {
			s := ly.b[j]
			for i := range ly.in {
				s += ly.w[i*ly.out+j] * cur[i]
			}
			z[j] = s
		}
		a := apply(z, ly.act)
		zs = append(zs, z)
		acts = append(acts, a)
		cur = a
	}
	return cur, acts, zs
}
