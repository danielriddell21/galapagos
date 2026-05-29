// Package neat implements NeuroEvolution of Augmenting Topologies: a
// population-based agent that evolves both the weights and the structure of its
// networks. It depends only on core, so it drops into any environment the
// genetic algorithm already runs in.
package neat

import (
	"math/rand/v2"
	"slices"

	"github.com/danielriddell21/galapagos/internal/core"
)

// Node types. Inputs feed observations, the bias node is constant 1, outputs
// produce the action vector, and hidden nodes are added by structural mutation.
const (
	nodeInput = iota
	nodeBias
	nodeOutput
	nodeHidden
)

// nodeGene identifies one neuron.
type nodeGene struct {
	id   int
	kind int
}

// connGene is a weighted, possibly disabled, connection between two nodes. The
// innovation number marks its historical origin so genomes align during
// crossover and distance computation.
type connGene struct {
	innovation int
	from, to   int
	weight     float64
	enabled    bool
}

// genome is a NEAT individual: a set of nodes and connections plus its measured
// fitness. Nodes are kept sorted by id and connections by innovation so all
// iteration is deterministic.
type genome struct {
	inputs, outputs int
	nodes           []nodeGene
	conns           []connGene
	fitness         core.Reward
	net             *network
}

// newMinimalGenome builds a fully connected perceptron: every input and the
// bias node wired to every output, with Gaussian weights. Hidden structure is
// added later by mutation.
func newMinimalGenome(inputs, outputs int, inno *innovations, rng *rand.Rand) *genome {
	g := &genome{inputs: inputs, outputs: outputs}
	for i := range inputs {
		g.nodes = append(g.nodes, nodeGene{id: i, kind: nodeInput})
	}
	biasID := inputs
	g.nodes = append(g.nodes, nodeGene{id: biasID, kind: nodeBias})
	for o := range outputs {
		outID := inputs + 1 + o
		g.nodes = append(g.nodes, nodeGene{id: outID, kind: nodeOutput})
		for in := 0; in <= inputs; in++ { // inputs plus the bias node
			g.conns = append(g.conns, connGene{
				innovation: inno.conn(in, outID),
				from:       in,
				to:         outID,
				weight:     rng.NormFloat64(),
				enabled:    true,
			})
		}
	}
	return g
}

// clone returns a deep copy with no shared backing arrays and an invalidated
// phenotype cache.
func (g *genome) clone() *genome {
	return &genome{
		inputs:  g.inputs,
		outputs: g.outputs,
		nodes:   slices.Clone(g.nodes),
		conns:   slices.Clone(g.conns),
	}
}

// nodeID returns the maximum node id in the genome, used when allocating new
// hidden nodes.
func (g *genome) maxNodeID() int {
	m := 0
	for _, n := range g.nodes {
		m = max(m, n.id)
	}
	return m
}

// hasNode reports whether a node with id exists.
func (g *genome) hasNode(id int) bool {
	for _, n := range g.nodes {
		if n.id == id {
			return true
		}
	}
	return false
}

// weights returns the connection weights as a flat slice, satisfying the
// core.Individual genome accessor. It is a lossy view (topology is omitted) used
// for display and inspection, not reconstruction.
func (g *genome) weights() []float64 {
	w := make([]float64, len(g.conns))
	for i, c := range g.conns {
		w[i] = c.weight
	}
	return w
}

// Act builds the phenotype on first use and returns its raw outputs, which the
// environment interprets. It implements core.Individual.
func (g *genome) Act(s core.State) core.Action {
	if g.net == nil {
		g.net = build(g)
	}
	return vecAction(g.net.forward(s.Observation()))
}

// Fitness implements core.Individual.
func (g *genome) Fitness() core.Reward { return g.fitness }

// SetFitness implements core.Individual.
func (g *genome) SetFitness(r core.Reward) { g.fitness = r }

// Genome implements core.Individual, returning the connection weights.
func (g *genome) Genome() []float64 { return g.weights() }

// vecAction carries a raw output vector as a core.Action.
type vecAction []float64

// Vector implements core.Action.
func (a vecAction) Vector() []float64 { return a }

var _ core.Individual = (*genome)(nil)
