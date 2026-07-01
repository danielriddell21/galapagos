package neat

import (
	"math/rand/v2"
	"slices"
)

func mutateWeights(g *genome, rate, std float64, rng *rand.Rand) {
	for i := range g.conns {
		if rng.Float64() >= rate {
			continue
		}
		if rng.Float64() < 0.1 {
			g.conns[i].weight = rng.NormFloat64()
		} else {
			g.conns[i].weight += rng.NormFloat64() * std
		}
	}
}

func (g *genome) hasConn(from, to int) bool {
	for _, c := range g.conns {
		if c.from == from && c.to == to {
			return true
		}
	}
	return false
}

func (g *genome) sources() []int {
	var ids []int
	for _, n := range g.nodes {
		if n.kind != nodeOutput {
			ids = append(ids, n.id)
		}
	}
	return ids
}

func (g *genome) targets() []int {
	var ids []int
	for _, n := range g.nodes {
		if n.kind == nodeHidden || n.kind == nodeOutput {
			ids = append(ids, n.id)
		}
	}
	return ids
}

func addConnection(g *genome, inno *innovations, rng *rand.Rand) {
	srcs, dsts := g.sources(), g.targets()
	if len(srcs) == 0 || len(dsts) == 0 {
		return
	}
	for range 20 {
		from := srcs[rng.IntN(len(srcs))]
		to := dsts[rng.IntN(len(dsts))]
		if from == to || g.hasConn(from, to) || reaches(g, to, from) {
			continue
		}
		g.conns = append(g.conns, connGene{
			innovation: inno.conn(from, to),
			from:       from,
			to:         to,
			weight:     rng.NormFloat64(),
			enabled:    true,
		})
		return
	}
}

func addNode(g *genome, inno *innovations, rng *rand.Rand) {
	enabled := enabledConns(g)
	if len(enabled) == 0 {
		return
	}
	ci := enabled[rng.IntN(len(enabled))]
	old := g.conns[ci]
	g.conns[ci].enabled = false

	newID := inno.newNode()
	for g.hasNode(newID) { // defensive: never collide with an existing id
		newID = inno.newNode()
	}
	g.nodes = append(g.nodes, nodeGene{id: newID, kind: nodeHidden})
	g.conns = append(g.conns,
		connGene{innovation: inno.conn(old.from, newID), from: old.from, to: newID, weight: 1, enabled: true},
		connGene{innovation: inno.conn(newID, old.to), from: newID, to: old.to, weight: old.weight, enabled: true},
	)
	slices.SortFunc(g.conns, func(a, b connGene) int { return a.innovation - b.innovation })
}

func enabledConns(g *genome) []int {
	var idx []int
	for i, c := range g.conns {
		if c.enabled {
			idx = append(idx, i)
		}
	}
	return idx
}
