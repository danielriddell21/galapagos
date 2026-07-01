package efficientcube

import (
	"math/rand/v2"

	rubix "github.com/danielriddell21/rubix/pkg/cube"

	"github.com/danielriddell21/galapagos/internal/nn"
)

const dataStream = 0x53c5ff8e3d9b1a77

type TrainConfig struct {
	Hidden []int
	K      int
	Batch  int
	Iters  int
	LR     float64
	Seed   int64
}

func DefaultTrainConfig() TrainConfig {
	return TrainConfig{Hidden: []int{512, 256}, K: 15, Batch: 1000, Iters: 20000, LR: 1e-3, Seed: 42}
}

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

func randomMove(rng *rand.Rand, prevFace int) rubix.Move {
	for {
		m := rubix.Move(rng.IntN(numMoves))
		if m.Face() != prevFace {
			return m
		}
	}
}

func accuracy(p *Policy, xs [][]float64, ys []int) float64 {
	correct := 0
	for i := range xs {
		if argmax(p.net.Forward(xs[i])) == ys[i] {
			correct++
		}
	}
	return float64(correct) / float64(len(xs))
}
