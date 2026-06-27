package chess

import (
	"math"

	gambit "github.com/danielriddell21/gambit/pkg/chess"

	"github.com/danielriddell21/galapagos/internal/core"
)

// bestLegalMove decodes an action into a legal move. The action is 128
// preferences (64 "from" + 64 "to") in the side-to-move's perspective; each
// legal move scores fromPref[from] + toPref[to], and the highest-scoring move
// wins. ok is false only when there are no legal moves (a terminal position).
func bestLegalMove(b *gambit.Board, a core.Action) (gambit.Move, bool) {
	moves := b.LegalMoves()
	if len(moves) == 0 {
		return gambit.NoMove, false
	}
	v := a.Vector()
	stm := b.SideToMove()

	best := moves[0]
	bestScore := math.Inf(-1)
	for _, m := range moves {
		score := pref(v, perspective(m.From(), stm)) + pref(v, 64+perspective(m.To(), stm))
		// Break ties toward queen promotion so the agent isn't penalised by
		// arbitrary underpromotions that share a from/to square.
		if m.IsPromotion() && m.Promotion() == gambit.Queen {
			score += 1e-6
		}
		if score > bestScore {
			bestScore, best = score, m
		}
	}
	return best, true
}

// pref reads index i from an action vector, treating out-of-range as zero so a
// short vector degrades gracefully.
func pref(v []float64, i int) float64 {
	if i >= 0 && i < len(v) {
		return v[i]
	}
	return 0
}
