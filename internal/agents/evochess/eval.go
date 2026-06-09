// Package evochess is an evolving, search-based chess agent. Each member carries
// an evolved board-evaluation function — material values plus piece-square tables
// — and chooses moves by a shallow alpha-beta search over the gambit engine,
// scoring leaf positions with that function. The genome is the flat vector of
// evaluation weights, evolved by a genetic algorithm. It depends only on core and
// the external gambit engine, never on an environment package.
//
// This is "approach B": where the ga/neat agents emit a move directly from a
// network, evochess evolves *judgement* (the evaluation) and pairs it with
// classical search, the way evolutionary chess engines (Blondie-style) do.
package evochess

import (
	gambit "github.com/danielriddell21/gambit/pkg/chess"
)

const (
	numTypes  = 7  // gambit.PieceType spans 0..6 (none, pawn..king)
	boardSize = 64 // squares
	// genomeLen is one material value per piece type plus a full piece-square
	// table per piece type.
	genomeLen = numTypes + numTypes*boardSize
)

// baseMaterial seeds evolution near the classical piece values (index by
// gambit.PieceType: none, pawn, knight, bishop, rook, queen, king). The king's
// value is irrelevant (it is never captured in legal play).
var baseMaterial = [numTypes]float64{0, 1, 3, 3, 5, 9, 0}

// evaluator scores a position from White's perspective: material plus a
// piece-square bonus read from each side's own point of view.
type evaluator struct {
	material [numTypes]float64
	pst      [numTypes][boardSize]float64
}

// newEvaluator decodes a genome into an evaluator.
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

// clone returns an independent copy (the arrays copy by value), so a frozen
// champion is unaffected by later evolution.
func (e *evaluator) clone() *evaluator {
	c := *e
	return &c
}

// score returns the position's value from White's perspective: positive favours
// White. Empty squares contribute nothing.
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
