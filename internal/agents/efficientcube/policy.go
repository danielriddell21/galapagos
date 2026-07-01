package efficientcube

import (
	"math"

	rubix "github.com/danielriddell21/rubix/pkg/cube"

	"github.com/danielriddell21/galapagos/internal/nn"
)

const inputDim = 54 * 6

const numMoves = int(rubix.NumMoves)

func encode(c rubix.Cube) []float64 {
	f := c.ToFacelets()
	x := make([]float64, inputDim)
	for i := range 54 {
		x[i*6+int(f[i])] = 1
	}
	return x
}

type Policy struct {
	net *nn.MLP
}

func NewPolicy(hidden []int, seed int64) *Policy {
	sizes := make([]int, 0, len(hidden)+2)
	sizes = append(sizes, inputDim)
	sizes = append(sizes, hidden...)
	sizes = append(sizes, numMoves)
	return &Policy{net: nn.New(nn.Config{Sizes: sizes, Hidden: nn.ReLU, Output: nn.Linear, Seed: seed})}
}

func (p *Policy) Probs(c rubix.Cube) []float64 { return nn.Softmax(p.net.Forward(encode(c))) }

func (p *Policy) logProbs(c rubix.Cube) []float64 {
	probs := p.Probs(c)
	for i := range probs {
		probs[i] = math.Log(max(probs[i], 1e-12))
	}
	return probs
}

func (p *Policy) Best(c rubix.Cube) rubix.Move {
	return rubix.Move(argmax(p.net.Forward(encode(c))))
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
