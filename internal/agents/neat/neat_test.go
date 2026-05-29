package neat

import (
	"math"
	"math/rand/v2"
	"testing"
)

func testConfig() Config {
	c := DefaultConfig()
	c.Population = 30
	c.Inputs = 8
	c.Outputs = 2
	c.Seed = 11
	return c
}

func TestMinimalGenomeForward(t *testing.T) {
	inno := newInnovations(2 + 1 + 1)
	g := newMinimalGenome(2, 1, inno, rand.New(rand.NewPCG(1, 2)))
	out := g.Act(obsState{0.5, -0.3}).Vector()
	if len(out) != 1 {
		t.Fatalf("output len = %d, want 1", len(out))
	}
	if math.IsNaN(out[0]) || math.IsInf(out[0], 0) {
		t.Fatalf("non-finite output %g", out[0])
	}
}

func TestInnovationNumbersReproducible(t *testing.T) {
	a := New(testConfig())
	b := New(testConfig())
	for i := range a.members {
		ca, cb := a.members[i].conns, b.members[i].conns
		if len(ca) != len(cb) {
			t.Fatalf("member %d connection count differs", i)
		}
		for j := range ca {
			if ca[j].innovation != cb[j].innovation {
				t.Fatalf("member %d conn %d innovation differs: %d vs %d", i, j, ca[j].innovation, cb[j].innovation)
			}
		}
	}
}

func TestMutationsStayAcyclic(t *testing.T) {
	inno := newInnovations(2 + 1 + 1)
	g := newMinimalGenome(2, 1, inno, rand.New(rand.NewPCG(3, 4)))
	rng := rand.New(rand.NewPCG(5, 6))
	for range 200 {
		switch rng.IntN(3) {
		case 0:
			addConnection(g, inno, rng)
		case 1:
			addNode(g, inno, rng)
		default:
			mutateWeights(g, 1, 0.5, rng)
		}
	}
	// A topological order covering every node exists only if the graph is acyclic.
	net := build(g)
	if len(net.order) != len(g.nodes) {
		t.Fatalf("topological order covers %d of %d nodes; graph has a cycle", len(net.order), len(g.nodes))
	}
}

func TestDistanceZeroForClone(t *testing.T) {
	g := New(testConfig()).members[0]
	if d := distance(g, g.clone(), 1, 1, 0.4); d != 0 {
		t.Fatalf("distance to clone = %g, want 0", d)
	}
}

func TestReachesDetectsPath(t *testing.T) {
	inno := newInnovations(2 + 1 + 1)
	g := newMinimalGenome(2, 1, inno, rand.New(rand.NewPCG(1, 1)))
	// Input 0 connects to the single output (id 3) in a minimal genome.
	if !reaches(g, 0, 3) {
		t.Fatal("expected a path from input 0 to output 3")
	}
	if reaches(g, 3, 0) {
		t.Fatal("there must be no path from output back to input")
	}
}

type obsState []float64

func (s obsState) Observation() []float64 { return s }
