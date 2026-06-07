package efficientcube

import (
	"slices"

	rubix "github.com/danielriddell21/rubix/pkg/cube"
)

// BeamConfig parameters beam search over the policy.
type BeamConfig struct {
	Width    int // beam width (candidates kept per depth)
	MaxDepth int // maximum solution length searched
}

// DefaultBeamConfig returns a balanced search configuration.
func DefaultBeamConfig() BeamConfig { return BeamConfig{Width: 1000, MaxDepth: 26} }

// Result is the outcome of a solve attempt.
type Result struct {
	Moves  []rubix.Move
	Solved bool
	Nodes  int
}

// beamNode is a partial solution: a cube, the moves taken to reach it, and the
// cumulative log-probability the policy assigned to those moves.
type beamNode struct {
	c     rubix.Cube
	path  []rubix.Move
	score float64
}

// Solve searches for a solution from start using beam search guided by the
// policy: at each depth it expands every beam node by all moves, scores children
// by cumulative policy log-probability, dedupes already-seen cubes, and keeps the
// top Width. It returns the first solution found or an unsolved result if the
// depth budget is exhausted.
func (p *Policy) Solve(start rubix.Cube, cfg BeamConfig) Result {
	if start.IsSolved() {
		return Result{Solved: true}
	}
	beam := []beamNode{{c: start}}
	visited := map[rubix.Cube]bool{start: true}
	nodes := 0

	for range cfg.MaxDepth {
		var cands []beamNode
		for _, bn := range beam {
			lp := p.logProbs(bn.c)
			for mi := range numMoves {
				child := bn.c.Applied(rubix.Move(mi))
				nodes++
				path := append(slices.Clone(bn.path), rubix.Move(mi))
				if child.IsSolved() {
					return Result{Moves: path, Solved: true, Nodes: nodes}
				}
				if visited[child] {
					continue
				}
				cands = append(cands, beamNode{c: child, path: path, score: bn.score + lp[mi]})
			}
		}
		if len(cands) == 0 {
			break
		}
		// Keep the highest-scoring candidates.
		slices.SortFunc(cands, func(a, b beamNode) int {
			switch {
			case a.score > b.score:
				return -1
			case a.score < b.score:
				return 1
			default:
				return 0
			}
		})
		if len(cands) > cfg.Width {
			cands = cands[:cfg.Width]
		}
		beam = cands
		for _, c := range beam {
			visited[c.c] = true
		}
	}
	return Result{Solved: false, Nodes: nodes}
}
