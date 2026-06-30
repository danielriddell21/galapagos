package neat

import (
	"math/rand/v2"
	"slices"

	"github.com/danielriddell21/galapagos/internal/core"
)

const (
	nodeInput = iota
	nodeBias
	nodeOutput
	nodeHidden
)

type nodeGene struct {
	id   int
	kind int
}

type connGene struct {
	innovation int
	from, to   int
	weight     float64
	enabled    bool
}

type genome struct {
	inputs, outputs int
	nodes           []nodeGene
	conns           []connGene
	fitness         core.Reward
	net             *network
}

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

func (g *genome) clone() *genome {
	return &genome{
		inputs:  g.inputs,
		outputs: g.outputs,
		nodes:   slices.Clone(g.nodes),
		conns:   slices.Clone(g.conns),
	}
}

func (g *genome) hasNode(id int) bool {
	for _, n := range g.nodes {
		if n.id == id {
			return true
		}
	}
	return false
}

func (g *genome) weights() []float64 {
	w := make([]float64, len(g.conns))
	for i, c := range g.conns {
		w[i] = c.weight
	}
	return w
}

func (g *genome) Act(s core.State) core.Action {
	if g.net == nil {
		g.net = build(g)
	}
	return vecAction(g.net.forward(s.Observation()))
}

func (g *genome) Fitness() core.Reward { return g.fitness }

func (g *genome) SetFitness(r core.Reward) { g.fitness = r }

func (g *genome) Genome() []float64 { return g.weights() }

type vecAction []float64

func (a vecAction) Vector() []float64 { return a }

var _ core.Individual = (*genome)(nil)
