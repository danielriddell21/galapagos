package sim

import (
	"math/rand/v2"
	"testing"
)

// draw collects n floats from r so streams can be compared.
func draw(r *rand.Rand, n int) []float64 {
	out := make([]float64, n)
	for i := range n {
		out[i] = r.Float64()
	}
	return out
}

func equal(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestStreamsReproducible(t *testing.T) {
	const seed = 42
	tests := []struct {
		name string
		make func() *rand.Rand
	}{
		{"master", func() *rand.Rand { return MasterRNG(seed) }},
		{"track", func() *rand.Rand { return TrackRNG(seed) }},
		{"member", func() *rand.Rand { return MemberRNG(seed, 7) }},
		{"gen", func() *rand.Rand { return GenRNG(seed, 3) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !equal(draw(tt.make(), 16), draw(tt.make(), 16)) {
				t.Fatalf("%s stream not reproducible for the same seed", tt.name)
			}
		})
	}
}

func TestMemberStreamsIndependentOfIndex(t *testing.T) {
	const seed = 42
	seen := map[float64]int{}
	for i := range 32 {
		first := MemberRNG(seed, i).Float64()
		if prev, ok := seen[first]; ok {
			t.Fatalf("member %d and %d produced identical streams", prev, i)
		}
		seen[first] = i
	}
}

func TestStreamsDifferFromEachOther(t *testing.T) {
	const seed = 42
	if equal(draw(MasterRNG(seed), 8), draw(TrackRNG(seed), 8)) {
		t.Fatal("master and track streams must not coincide")
	}
	if equal(draw(GenRNG(seed, 0), 8), draw(MemberRNG(seed, 0), 8)) {
		t.Fatal("gen and member streams must not coincide")
	}
}
