package neat

import (
	"math/rand/v2"
	"slices"
)

// mutateWeights perturbs each connection weight with Gaussian noise at the
// given rate, occasionally replacing a weight outright.
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

// hasConn reports whether a from→to connection already exists.
func (g *genome) hasConn(from, to int) bool {
	for _, c := range g.conns {
		if c.from == from && c.to == to {
			return true
		}
	}
	return false
}

// sources returns ids that may originate a connection: inputs, the bias node,
// and hidden nodes.
func (g *genome) sources() []int {
	var ids []int
	for _, n := range g.nodes {
		if n.kind != nodeOutput {
			ids = append(ids, n.id)
		}
	}
	return ids
}

// targets returns ids that may receive a connection: hidden and output nodes.
func (g *genome) targets() []int {
	var ids []int
	for _, n := range g.nodes {
		if n.kind == nodeHidden || n.kind == nodeOutput {
			ids = append(ids, n.id)
		}
	}
	return ids
}

// addConnection adds a new acyclic connection between two unconnected nodes, if
// a valid pair is found within a few attempts.
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

// addNode splits a random enabled connection in two with a new hidden node, the
// classic NEAT structural mutation that preserves behavior initially.
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

// enabledConns returns the indices of enabled connections.
func enabledConns(g *genome) []int {
	var idx []int
	for i, c := range g.conns {
		if c.enabled {
			idx = append(idx, i)
		}
	}
	return idx
}
