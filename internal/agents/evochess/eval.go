package evochess

import (
	gambit "github.com/danielriddell21/gambit/pkg/chess"
)

const (
	numTypes  = 7
	boardSize = 64

	genomeLen = numTypes + numTypes*boardSize
)

var baseMaterial = [numTypes]float64{0, 1, 3, 3, 5, 9, 0}

type evaluator struct {
	material [numTypes]float64
	pst      [numTypes][boardSize]float64
}

func newEvaluator(genome []float64) *evaluator {
	e := &evaluator{}
	for t := range numTypes {
		e.material[t] = genome[t]
	}
	off := numTypes
	for t := range numTypes {
		for sq := range boardSize {
			e.pst[t][sq] = genome[off]
			off++
		}
	}
	return e
}

func (e *evaluator) clone() *evaluator {
	c := *e
	return &c
}

func (e *evaluator) score(b *gambit.Board) float64 {
	var sum float64
	b.Each(func(s gambit.Square, p gambit.Piece) {
		if p.IsEmpty() {
			return
		}
		t := p.Type()
		idx := int(s)
		if p.Color() == gambit.Black {
			idx ^= 56 // read the table from Black's own perspective
		}
		v := e.material[t] + e.pst[t][idx]
		if p.Color() == gambit.White {
			sum += v
		} else {
			sum -= v
		}
	})
	return sum
}
