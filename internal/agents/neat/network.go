package neat

import (
	"math"
	"slices"
)

// network is the phenotype compiled from a genome: nodes evaluated in
// topological order with tanh activations. NEAT genomes are kept acyclic, so a
// single forward pass suffices.
type network struct {
	inputs, outputs int
	biasID          int
	order           []int
	incoming        map[int][]connGene
	kind            map[int]int
}

// build compiles a genome into an evaluable network.
func build(g *genome) *network {
	kind := make(map[int]int, len(g.nodes))
	for _, n := range g.nodes {
		kind[n.id] = n.kind
	}
	incoming := map[int][]connGene{}
	indeg := map[int]int{}
	adj := map[int][]int{}
	for _, n := range g.nodes {
		indeg[n.id] = 0
	}
	for _, c := range g.conns {
		if !c.enabled {
			continue
		}
		incoming[c.to] = append(incoming[c.to], c)
		adj[c.from] = append(adj[c.from], c.to)
		indeg[c.to]++
	}

	// Kahn's algorithm with a sorted frontier for deterministic ordering.
	var frontier []int
	for id, d := range indeg {
		if d == 0 {
			frontier = append(frontier, id)
		}
	}
	slices.Sort(frontier)
	var order []int
	for len(frontier) > 0 {
		id := frontier[0]
		frontier = frontier[1:]
		order = append(order, id)
		next := slices.Clone(adj[id])
		slices.Sort(next)
		for _, to := range next {
			indeg[to]--
			if indeg[to] == 0 {
				frontier = append(frontier, to)
				slices.Sort(frontier)
			}
		}
	}

	return &network{
		inputs:   g.inputs,
		outputs:  g.outputs,
		biasID:   g.inputs,
		order:    order,
		incoming: incoming,
		kind:     kind,
	}
}

// forward evaluates the network for an observation and returns the raw output
// values.
func (n *network) forward(obs []float64) []float64 {
	val := make(map[int]float64, len(n.order))
	for i := range n.inputs {
		val[i] = obs[i]
	}
	val[n.biasID] = 1
	for _, id := range n.order {
		switch n.kind[id] {
		case nodeInput, nodeBias:
			continue
		}
		var sum float64
		for _, c := range n.incoming[id] {
			sum += c.weight * val[c.from]
		}
		val[id] = math.Tanh(sum)
	}
	out := make([]float64, n.outputs)
	for o := range n.outputs {
		out[o] = val[n.inputs+1+o]
	}
	return out
}

// reaches reports whether dst is reachable from src by following enabled
// connections, used to keep the network acyclic when adding connections.
func reaches(g *genome, src, dst int) bool {
	if src == dst {
		return true
	}
	adj := map[int][]int{}
	for _, c := range g.conns {
		if c.enabled {
			adj[c.from] = append(adj[c.from], c.to)
		}
	}
	seen := map[int]bool{src: true}
	stack := []int{src}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, to := range adj[cur] {
			if to == dst {
				return true
			}
			if !seen[to] {
				seen[to] = true
				stack = append(stack, to)
			}
		}
	}
	return false
}
