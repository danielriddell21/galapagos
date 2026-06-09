package evochess

import (
	"math"

	"github.com/danielriddell21/galapagos/internal/core"
	gambit "github.com/danielriddell21/gambit/pkg/chess"
)

// mateScore is the magnitude returned for a decided game, far larger than any
// material evaluation so checkmates dominate the search.
const mateScore = 1e6

// search returns the minimax value of b to depth, from White's perspective,
// with alpha-beta pruning. White maximizes, Black minimizes.
func (e *evaluator) search(b *gambit.Board, depth int, alpha, beta float64) float64 {
	if res, _ := b.Status(); res != gambit.InProgress {
		switch res {
		case gambit.WhiteWins:
			return mateScore
		case gambit.BlackWins:
			return -mateScore
		default:
			return 0
		}
	}
	if depth == 0 {
		return e.score(b)
	}
	moves := b.LegalMoves()
	if b.SideToMove() == gambit.White {
		val := math.Inf(-1)
		for _, m := range moves {
			val = math.Max(val, e.search(b.ApplyMove(m), depth-1, alpha, beta))
			alpha = math.Max(alpha, val)
			if alpha >= beta {
				break
			}
		}
		return val
	}
	val := math.Inf(1)
	for _, m := range moves {
		val = math.Min(val, e.search(b.ApplyMove(m), depth-1, alpha, beta))
		beta = math.Min(beta, val)
		if alpha >= beta {
			break
		}
	}
	return val
}

// bestMove returns the side-to-move's best move under an alpha-beta search of the
// given depth, resolving ties to the first move for determinism. It assumes b is
// not already terminal.
func bestMove(e *evaluator, b *gambit.Board, depth int) gambit.Move {
	moves := b.LegalMoves()
	if len(moves) == 0 {
		return gambit.NoMove
	}
	white := b.SideToMove() == gambit.White
	best := moves[0]
	bestVal := math.Inf(-1)
	if !white {
		bestVal = math.Inf(1)
	}
	for _, m := range moves {
		v := e.search(b.ApplyMove(m), depth-1, math.Inf(-1), math.Inf(1))
		if (white && v > bestVal) || (!white && v < bestVal) {
			bestVal, best = v, m
		}
	}
	return best
}

// action carries a move as a from/to preference vector the chess environment
// decodes back into that move.
type action []float64

func (a action) Vector() []float64 { return a }

// encodeMove turns a chosen move into the chess environment's 128-element action
// (64 "from" + 64 "to" preferences) in the side-to-move's perspective, spiking
// the chosen move's squares so the environment decodes back to exactly this move.
func encodeMove(b *gambit.Board, m gambit.Move) core.Action {
	v := make(action, 128)
	if m == gambit.NoMove {
		return v
	}
	stm := b.SideToMove()
	v[perspective(m.From(), stm)] = 1
	v[64+perspective(m.To(), stm)] = 1
	return v
}

// perspective maps a square to its index from c's point of view, matching the
// chess environment's action decoding: identity for White, rank-mirrored for
// Black.
func perspective(s gambit.Square, c gambit.Color) int {
	if c == gambit.White {
		return int(s)
	}
	return int(s) ^ 56
}
