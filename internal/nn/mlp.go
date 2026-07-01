package nn

import (
	"math"
	"math/rand/v2"
	"slices"
)

type Activation uint8

const (
	Linear Activation = iota

	ReLU
)

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

func deriv(z float64, a Activation) float64 {
	if a == ReLU {
		if z > 0 {
			return 1
		}
		return 0
	}
	return 1
}

type layer struct {
	in, out int
	w, b    []float64
	act     Activation
}

type MLP struct {
	layers []layer
	sizes  []int
	hidden Activation
	output Activation
}

type Config struct {
	Sizes  []int
	Hidden Activation
	Output Activation
	Seed   int64
}

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

func (m *MLP) InDim() int { return m.sizes[0] }

func (m *MLP) OutDim() int { return m.sizes[len(m.sizes)-1] }

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

func (m *MLP) forwardCached(x []float64) (out []float64, acts, zs [][]float64) {
	acts = make([][]float64, 1, 1+len(m.layers))
	acts[0] = x
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
