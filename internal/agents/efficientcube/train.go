package efficientcube

import (
	"math/rand/v2"

	rubix "github.com/danielriddell21/rubix/pkg/cube"

	"github.com/danielriddell21/galapagos/internal/nn"
)

// dataStream decorrelates the data-generation RNG from weight initialization.
const dataStream = 0x53c5ff8e3d9b1a77

// TrainConfig parameters EfficientCube training.
type TrainConfig struct {
	Hidden []int   // hidden layer sizes
	K      int     // maximum scramble length (training distribution depth)
	Batch  int     // examples per optimizer step
	Iters  int     // optimizer steps
	LR     float64 // Adam learning rate
	Seed   int64
}

// DefaultTrainConfig returns balanced defaults for moderate scrambles.
func DefaultTrainConfig() TrainConfig {
	return TrainConfig{Hidden: []int{512, 256}, K: 15, Batch: 1000, Iters: 20000, LR: 1e-3, Seed: 42}
}

// Train learns a policy by self-supervision: it scrambles the solved cube and
// trains the network to predict, at each state, the move that reverses the last
// scramble step (a descent move toward solved). logFn, if non-nil, receives the
// iteration, loss, and batch move-prediction accuracy periodically. Training is
// deterministic for a given config and seed.
func Train(cfg TrainConfig, logFn func(iter int, loss, acc float64)) *Policy {
	p := NewPolicy(cfg.Hidden, cfg.Seed)
	opt := nn.NewAdam(p.net, cfg.LR)
	rng := rand.New(rand.NewPCG(uint64(cfg.Seed)^dataStream, 0))

	for it := range cfg.Iters {
		xs := make([][]float64, 0, cfg.Batch)
		ys := make([]int, 0, cfg.Batch)
		for len(xs) < cfg.Batch {
			genSequence(rng, cfg.K, cfg.Batch, &xs, &ys)
		}
		loss := p.net.TrainBatch(opt, xs, ys)
		if logFn != nil && it%64 == 0 {
			logFn(it, loss, accuracy(p, xs, ys))
		}
	}
	return p
}

// genSequence applies a random scramble and appends (state, reversing-move)
// examples until the batch is full.
func genSequence(rng *rand.Rand, k, batch int, xs *[][]float64, ys *[]int) {
	length := 1 + rng.IntN(k)
	c := rubix.Solved()
	prevFace := -1
	for t := 0; t < length && len(*xs) < batch; t++ {
		m := randomMove(rng, prevFace)
		c = c.Applied(m)
		// The move that returns toward solved from the new state is m's inverse.
		*xs = append(*xs, encode(c))
		*ys = append(*ys, int(m.Inverse()))
		prevFace = m.Face()
	}
}

// randomMove picks a move whose face differs from the previous one, avoiding
// trivially redundant or cancelling consecutive turns.
func randomMove(rng *rand.Rand, prevFace int) rubix.Move {
	for {
		m := rubix.Move(rng.IntN(numMoves))
		if m.Face() != prevFace {
			return m
		}
	}
}

// accuracy reports the fraction of a batch whose argmax move matches the label.
func accuracy(p *Policy, xs [][]float64, ys []int) float64 {
	correct := 0
	for i := range xs {
		if argmax(p.net.Forward(xs[i])) == ys[i] {
			correct++
		}
	}
	return float64(correct) / float64(len(xs))
}
