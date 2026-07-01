package evochess

import (
	"math"

	gambit "github.com/danielriddell21/gambit/pkg/chess"

	"github.com/danielriddell21/galapagos/internal/core"
)

const mateScore = 1e6

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

type action []float64

func (a action) Vector() []float64 { return a }

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

func perspective(s gambit.Square, c gambit.Color) int {
	if c == gambit.White {
		return int(s)
	}
	return int(s) ^ 56
}
