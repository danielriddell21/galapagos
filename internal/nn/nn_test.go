package nn

import (
	"math"
	"math/rand/v2"
	"path/filepath"
	"slices"
	"testing"
)

func TestGradientCheck(t *testing.T) {
	m := New(Config{Sizes: []int{4, 5, 3}, Hidden: ReLU, Output: Linear, Seed: 1})
	rng := rand.New(rand.NewPCG(2, 3))
	x := make([]float64, 4)
	for i := range x {
		x[i] = rng.NormFloat64()
	}
	const target = 1

	out, acts, zs := m.forwardCached(x)
	_, dOut := SoftmaxCrossEntropy(out, target)
	g := m.newGrads()
	m.backwardInto(g, acts, zs, dOut)

	const eps = 1e-6
	loss := func() float64 {
		l, _ := SoftmaxCrossEntropy(m.Forward(x), target)
		return l
	}
	for l := range m.layers {
		for i := range m.layers[l].w {
			orig := m.layers[l].w[i]
			m.layers[l].w[i] = orig + eps
			lp := loss()
			m.layers[l].w[i] = orig - eps
			lm := loss()
			m.layers[l].w[i] = orig
			num := (lp - lm) / (2 * eps)
			ana := g.dW[l][i]
			if math.Abs(num-ana) > 1e-4*(1+math.Abs(ana)) {
				t.Fatalf("layer %d weight %d: analytic %g vs numeric %g", l, i, ana, num)
			}
		}
	}
}

func TestLearnsClassification(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 9))
	// class 0 around (-1,-1), class 1 around (+1,+1)
	makeBatch := func(n int) (xs [][]float64, ys []int) {
		for range n {
			c := rng.IntN(2)
			cx := -1.0
			if c == 1 {
				cx = 1.0
			}
			xs = append(xs, []float64{cx + 0.3*rng.NormFloat64(), cx + 0.3*rng.NormFloat64()})
			ys = append(ys, c)
		}
		return xs, ys
	}
	m := New(Config{Sizes: []int{2, 16, 2}, Hidden: ReLU, Output: Linear, Seed: 5})
	opt := NewAdam(m, 0.01)
	var first, last float64
	for it := range 400 {
		xs, ys := makeBatch(64)
		l := m.TrainBatch(opt, xs, ys)
		if it == 0 {
			first = l
		}
		last = l
	}
	if last >= first {
		t.Fatalf("loss did not fall: first %g last %g", first, last)
	}
	// Accuracy on a fresh batch.
	xs, ys := makeBatch(200)
	correct := 0
	for i := range xs {
		if argmax(m.Forward(xs[i])) == ys[i] {
			correct++
		}
	}
	if correct < 190 {
		t.Fatalf("accuracy %d/200 too low", correct)
	}
}

func TestSeedDeterminism(t *testing.T) {
	a := New(Config{Sizes: []int{3, 4, 2}, Hidden: ReLU, Output: Linear, Seed: 42})
	b := New(Config{Sizes: []int{3, 4, 2}, Hidden: ReLU, Output: Linear, Seed: 42})
	for l := range a.layers {
		if !slices.Equal(a.layers[l].w, b.layers[l].w) {
			t.Fatalf("layer %d weights differ for identical seed", l)
		}
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	m := New(Config{Sizes: []int{5, 6, 4}, Hidden: ReLU, Output: Linear, Seed: 11})
	path := filepath.Join(t.TempDir(), "net.json")
	if err := m.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	x := []float64{0.1, -0.2, 0.3, 0.4, -0.5}
	if !slices.Equal(m.Forward(x), got.Forward(x)) {
		t.Fatal("loaded network produces different output")
	}
}

func argmax(v []float64) int {
	best := 0
	for i, x := range v {
		if x > v[best] {
			best = i
		}
	}
	return best
}
