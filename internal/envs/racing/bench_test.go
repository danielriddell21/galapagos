package racing

import (
	"math/rand/v2"
	"testing"

	"github.com/danielriddell21/galapagos/internal/core"
)

func BenchmarkGenerateTrack(b *testing.B) {
	p := DefaultTrackParams()
	for i := 0; b.Loop(); i++ {
		GenerateTrack(rand.New(rand.NewPCG(uint64(i), 0x9e37)), p)
	}
}

func BenchmarkStepAll(b *testing.B) {
	e := New(DefaultConfig())
	const n = 100
	e.ResetAll(n, rand.New(rand.NewPCG(1, 2)))
	actions := make([]core.Action, n)
	for i := range actions {
		actions[i] = NewAction(0.2, 1)
	}
	b.ResetTimer()
	for b.Loop() {
		if e.Alive() == 0 {
			e.ResetAll(n, rand.New(rand.NewPCG(1, 2)))
		}
		e.StepAll(actions)
	}
}
