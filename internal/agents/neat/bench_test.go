package neat

import (
	"testing"

	"github.com/danielriddell21/galapagos/internal/core"
)

func BenchmarkForward(b *testing.B) {
	g := New(testConfig()).members[0]
	x := obsState(make([]float64, testConfig().Inputs))
	for b.Loop() {
		g.net = nil // rebuild each call to also measure phenotype compilation
		g.Act(x)
	}
}

func BenchmarkEvolve(b *testing.B) {
	p := New(testConfig())
	for b.Loop() {
		for i, m := range p.members {
			m.fitness = core.Reward(i)
		}
		p.Evolve()
	}
}
