package neat

import (
	"math"
	"math/rand/v2"
	"slices"
)

// connsByInnovation indexes a genome's connections by innovation number.
func connsByInnovation(g *genome) map[int]connGene {
	m := make(map[int]connGene, len(g.conns))
	for _, c := range g.conns {
		m[c.innovation] = c
	}
	return m
}

// distance computes the NEAT compatibility distance between two genomes from
// the number of excess and disjoint genes and the average weight difference of
// matching genes. It is the metric used to group genomes into species.
func distance(a, b *genome, c1, c2, c3 float64) float64 {
	ma, mb := connsByInnovation(a), connsByInnovation(b)
	maxA := maxInnovation(a)
	maxB := maxInnovation(b)
	border := min(maxA, maxB)

	var disjoint, excess, matching int
	var weightDiff float64
	seen := map[int]bool{}
	for inno, ca := range ma {
		seen[inno] = true
		if cb, ok := mb[inno]; ok {
			matching++
			weightDiff += math.Abs(ca.weight - cb.weight)
		} else if inno <= border {
			disjoint++
		} else {
			excess++
		}
	}
	for inno := range mb {
		if seen[inno] {
			continue
		}
		if inno <= border {
			disjoint++
		} else {
			excess++
		}
	}

	n := max(max(len(a.conns), len(b.conns)), 1)
	avgWeight := 0.0
	if matching > 0 {
		avgWeight = weightDiff / float64(matching)
	}
	return c1*float64(excess)/float64(n) + c2*float64(disjoint)/float64(n) + c3*avgWeight
}

func maxInnovation(g *genome) int {
	m := 0
	for _, c := range g.conns {
		m = max(m, c.innovation)
	}
	return m
}

// crossover breeds a child from two parents. Matching genes are inherited from
// a random parent; disjoint and excess genes come from the fitter parent
// (parentA is assumed at least as fit as parentB).
func crossover(a, b *genome, rng *rand.Rand) *genome {
	mb := connsByInnovation(b)
	child := &genome{inputs: a.inputs, outputs: a.outputs}

	nodeSet := map[int]nodeGene{}
	addNodes := func(g *genome) {
		for _, n := range g.nodes {
			if n.kind != nodeHidden {
				nodeSet[n.id] = n
			}
		}
	}
	addNodes(a)
	addNodes(b)

	for _, ca := range a.conns {
		gene := ca
		if cb, ok := mb[ca.innovation]; ok && rng.IntN(2) == 0 {
			gene = cb
		}
		// A disabled gene in either parent may re-enable in the child.
		if !gene.enabled && rng.Float64() < 0.75 {
			gene.enabled = false
		} else {
			gene.enabled = true
		}
		child.conns = append(child.conns, gene)
		ensureHidden(nodeSet, a, b, gene.from)
		ensureHidden(nodeSet, a, b, gene.to)
	}

	for _, n := range nodeSet {
		child.nodes = append(child.nodes, n)
	}
	slices.SortFunc(child.nodes, func(x, y nodeGene) int { return x.id - y.id })
	slices.SortFunc(child.conns, func(x, y connGene) int { return x.innovation - y.innovation })
	return child
}

// ensureHidden adds a hidden node referenced by an inherited connection, taking
// its definition from whichever parent declares it.
func ensureHidden(set map[int]nodeGene, a, b *genome, id int) {
	if _, ok := set[id]; ok {
		return
	}
	for _, g := range []*genome{a, b} {
		for _, n := range g.nodes {
			if n.id == id {
				set[id] = n
				return
			}
		}
	}
}
