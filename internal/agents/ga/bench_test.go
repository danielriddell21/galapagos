package ga

import (
	"math/rand/v2"
	"testing"
)

func BenchmarkForward(b *testing.B) {
	cfg := testConfig()
	rng := rand.New(rand.NewPCG(1, 2))
	n := newNet(cfg.Inputs, cfg.HiddenSize, cfg.Outputs, randomGenome(GenomeLen(cfg.Inputs, cfg.HiddenSize, cfg.Outputs), rng))
	x := make([]float64, cfg.Inputs)
	for b.Loop() {
		n.forward(x)
	}
}

func BenchmarkEvolve(b *testing.B) {
	p := New(testConfig())
	for b.Loop() {
		assignFitness(p)
		p.Evolve()
	}
}
