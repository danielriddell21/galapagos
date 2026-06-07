// Package efficientcube implements EfficientCube (Takano, TMLR 2023): a
// self-supervised policy that learns to solve the Rubik's cube. It scrambles the
// solved cube and trains a classifier to predict the move that reverses each
// scramble step, then solves new scrambles with beam search over the policy. It
// needs no external solver — only the public rubix cube model — and depends only
// on core, nn, and rubix (never an environment package).
package efficientcube

import (
	"math"

	"github.com/danielriddell21/galapagos/internal/nn"
	rubix "github.com/danielriddell21/rubix/pkg/cube"
)

// inputDim is the one-hot encoding size: 54 facelets × 6 colors.
const inputDim = 54 * 6

// numMoves is the number of distinct cube moves (policy output size).
const numMoves = int(rubix.NumMoves)

// encode returns the one-hot facelet encoding of a cube.
func encode(c rubix.Cube) []float64 {
	f := c.ToFacelets()
	x := make([]float64, inputDim)
	for i := range 54 {
		x[i*6+int(f[i])] = 1
	}
	return x
}

// Policy is a move-prediction network over cube states.
type Policy struct {
	net *nn.MLP
}

// NewPolicy builds a policy network with the given hidden layer sizes.
func NewPolicy(hidden []int, seed int64) *Policy {
	sizes := make([]int, 0, len(hidden)+2)
	sizes = append(sizes, inputDim)
	sizes = append(sizes, hidden...)
	sizes = append(sizes, numMoves)
	return &Policy{net: nn.New(nn.Config{Sizes: sizes, Hidden: nn.ReLU, Output: nn.Linear, Seed: seed})}
}

// Probs returns the move probability distribution for a state.
func (p *Policy) Probs(c rubix.Cube) []float64 { return nn.Softmax(p.net.Forward(encode(c))) }

// logProbs returns log move probabilities (for beam-search scoring).
func (p *Policy) logProbs(c rubix.Cube) []float64 {
	probs := p.Probs(c)
	for i := range probs {
		probs[i] = math.Log(max(probs[i], 1e-12))
	}
	return probs
}

// Best returns the highest-probability move for a state.
func (p *Policy) Best(c rubix.Cube) rubix.Move {
	return rubix.Move(argmax(p.net.Forward(encode(c))))
}

// argmax returns the index of the largest value (ties to the lowest index).
func argmax(v []float64) int {
	best := 0
	for i, x := range v {
		if x > v[best] {
			best = i
		}
	}
	return best
}
