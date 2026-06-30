package efficientcube

import (
	"slices"

	rubix "github.com/danielriddell21/rubix/pkg/cube"
)

type BeamConfig struct {
	Width    int
	MaxDepth int
}

func DefaultBeamConfig() BeamConfig { return BeamConfig{Width: 1000, MaxDepth: 26} }

type Result struct {
	Moves  []rubix.Move
	Solved bool
	Nodes  int
}

type beamNode struct {
	c     rubix.Cube
	path  []rubix.Move
	score float64
}

func (p *Policy) Solve(start rubix.Cube, cfg BeamConfig) Result {
	if start.IsSolved() {
		return Result{Solved: true}
	}
	beam := []beamNode{{c: start}}
	visited := map[rubix.Cube]bool{start: true}
	nodes := 0

	for range cfg.MaxDepth {
		cands, sol, found := p.expandBeam(beam, visited, &nodes)
		if found {
			return sol
		}
		if len(cands) == 0 {
			break
		}
		beam = topCandidates(cands, cfg.Width)
		for _, c := range beam {
			visited[c.c] = true
		}
	}
	return Result{Solved: false, Nodes: nodes}
}

func (p *Policy) expandBeam(beam []beamNode, visited map[rubix.Cube]bool, nodes *int) ([]beamNode, Result, bool) {
	var cands []beamNode
	for _, bn := range beam {
		lp := p.logProbs(bn.c)
		for mi := range numMoves {
			child := bn.c.Applied(rubix.Move(mi))
			*nodes++
			path := append(slices.Clone(bn.path), rubix.Move(mi))
			if child.IsSolved() {
				return nil, Result{Moves: path, Solved: true, Nodes: *nodes}, true
			}
			if visited[child] {
				continue
			}
			cands = append(cands, beamNode{c: child, path: path, score: bn.score + lp[mi]})
		}
	}
	return cands, Result{}, false
}

func topCandidates(cands []beamNode, width int) []beamNode {
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
	if len(cands) > width {
		cands = cands[:width]
	}
	return cands
}
