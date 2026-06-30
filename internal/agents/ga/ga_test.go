package ga

import (
	"math"
	"math/rand/v2"
	"path/filepath"
	"slices"
	"testing"

	"github.com/danielriddell21/galapagos/internal/core"
)

func testConfig() Config {
	return Config{
		Population:    20,
		EliteFraction: 0.1,
		MutationRate:  0.05,
		MutationStd:   0.2,
		HiddenSize:    8,
		Inputs:        8,
		Outputs:       2,
		Seed:          42,
	}
}

func TestGenomeLen(t *testing.T) {
	// 8*8 + 8 + 8*2 + 2 = 64 + 8 + 16 + 2 = 90
	if got := GenomeLen(8, 8, 2); got != 90 {
		t.Fatalf("GenomeLen = %d, want 90", got)
	}
}

func TestNetForwardShapeAndRange(t *testing.T) {
	cfg := testConfig()
	rng := rand.New(rand.NewPCG(1, 2))
	g := randomGenome(GenomeLen(cfg.Inputs, cfg.HiddenSize, cfg.Outputs), rng)
	n := newNet(cfg.Inputs, cfg.HiddenSize, cfg.Outputs, g)
	out := n.forward(make([]float64, cfg.Inputs))
	if len(out) != cfg.Outputs {
		t.Fatalf("output len = %d, want %d", len(out), cfg.Outputs)
	}
	// With zero input, output equals the output biases (finite).
	for _, v := range out {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Fatalf("non-finite output %g", v)
		}
	}
}

func TestActionShapeAndFinite(t *testing.T) {
	cfg := testConfig()
	rng := rand.New(rand.NewPCG(5, 9))
	m := newIndividual(cfg, randomGenome(GenomeLen(cfg.Inputs, cfg.HiddenSize, cfg.Outputs), rng))
	for range 100 {
		obs := make([]float64, cfg.Inputs)
		for i := range obs {
			obs[i] = rng.Float64()
		}
		// The agent emits raw outputs (one per network output); the environment
		// interprets them. They must be finite and correctly shaped.
		v := m.Act(fakeState(obs)).Vector()
		if len(v) != cfg.Outputs {
			t.Fatalf("action len = %d, want %d", len(v), cfg.Outputs)
		}
		for _, x := range v {
			if math.IsNaN(x) || math.IsInf(x, 0) {
				t.Fatalf("non-finite action component %g", x)
			}
		}
	}
}

func TestPopulationReproducible(t *testing.T) {
	a := New(testConfig())
	b := New(testConfig())
	ia, ib := collect(a), collect(b)
	for i := range ia {
		if !slices.Equal(ia[i].Genome(), ib[i].Genome()) {
			t.Fatalf("member %d genome differs between identical seeds", i)
		}
	}
}

func TestEvolvePreservesEliteAndIsDeterministic(t *testing.T) {
	p1 := New(testConfig())
	assignFitness(p1)
	bestBefore := slices.Clone(p1.Best())
	p1.Evolve()
	// The top elite genome must survive unmutated as some member.
	if !containsGenome(p1, bestBefore) {
		t.Fatal("elite genome did not survive evolution unmutated")
	}

	// A second identically-seeded run must evolve identically.
	p2 := New(testConfig())
	assignFitness(p2)
	p2.Evolve()
	for i, m := range collect(p1) {
		if !slices.Equal(m.Genome(), collect(p2)[i].Genome()) {
			t.Fatalf("evolution not deterministic at member %d", i)
		}
	}
}

func TestNewSetsDiversityDefaults(t *testing.T) {
	// Leaving the diversity fields zero must fill in sensible defaults so every
	// GA run resists premature convergence without explicit configuration.
	p := New(testConfig())
	if p.cfg.ImmigrantFraction <= 0 || p.cfg.StagnationWindow <= 0 || p.cfg.HyperMutation <= 1 {
		t.Fatalf("diversity defaults not set: imm=%v win=%d hyper=%v",
			p.cfg.ImmigrantFraction, p.cfg.StagnationWindow, p.cfg.HyperMutation)
	}
}

func TestEvolveInjectsImmigrants(t *testing.T) {
	p := New(testConfig())
	assignFitness(p)
	n := len(p.members)
	eliteCount := max(1, int(p.cfg.EliteFraction*float64(n)))
	immigrantCount := int(p.cfg.ImmigrantFraction * float64(n))
	if immigrantCount == 0 {
		t.Fatal("expected at least one immigrant per generation")
	}
	p.Evolve()

	// The immigrant slots (just after the elites) must be the fresh random
	// genomes drawn from the dedicated immigrant stream for generation 0.
	genomeLen := GenomeLen(p.cfg.Inputs, p.cfg.HiddenSize, p.cfg.Outputs)
	members := collect(p)
	for i := range immigrantCount {
		irng := rand.New(rand.NewPCG(uint64(p.cfg.Seed)^streamImmigrant, uint64(i)))
		want := randomGenome(genomeLen, irng)
		if !slices.Equal(members[eliteCount+i].Genome(), want) {
			t.Fatalf("immigrant %d is not the expected fresh random genome", i)
		}
	}
}

func TestStagnationWidensMutation(t *testing.T) {
	p := New(testConfig())
	// Hold the best fitness flat across generations so the champion never
	// improves; the stagnation counter must climb past the window, which is what
	// triggers the widened (hyper) mutation that breaks out of a local optimum.
	for range p.cfg.StagnationWindow + 2 {
		for _, m := range collect(p) {
			m.SetFitness(5)
		}
		p.Evolve()
	}
	if p.stalled < p.cfg.StagnationWindow {
		t.Fatalf("stalled = %d, want >= window %d after flat fitness", p.stalled, p.cfg.StagnationWindow)
	}

	// A genuine improvement must reset the stagnation counter.
	for _, m := range collect(p) {
		m.SetFitness(100)
	}
	p.Evolve()
	if p.stalled != 0 {
		t.Fatalf("stalled = %d, want 0 after the champion improved", p.stalled)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	p := New(testConfig())
	assignFitness(p)
	path := filepath.Join(t.TempDir(), "best.json")
	if err := p.SaveBest(path); err != nil {
		t.Fatalf("SaveBest: %v", err)
	}
	sg, err := LoadGenome(path)
	if err != nil {
		t.Fatalf("LoadGenome: %v", err)
	}
	if !slices.Equal(sg.Genome, p.Best()) {
		t.Fatal("loaded genome differs from saved best")
	}
	d := NewDriver(sg)
	if got := len(d.Act(fakeState(make([]float64, sg.Inputs))).Vector()); got != sg.Outputs {
		t.Fatalf("driver action len = %d, want %d", got, sg.Outputs)
	}
}

func TestSetMemberGenome(t *testing.T) {
	p := New(testConfig())
	g := make([]float64, GenomeLen(p.cfg.Inputs, p.cfg.HiddenSize, p.cfg.Outputs))
	for i := range g {
		g[i] = 0.5
	}
	if err := p.SetMemberGenome(0, g); err != nil {
		t.Fatalf("SetMemberGenome: %v", err)
	}
	if !slices.Equal(collect(p)[0].Genome(), g) {
		t.Fatal("member 0 genome not replaced")
	}
	if err := p.SetMemberGenome(0, g[:3]); err == nil {
		t.Fatal("expected error for wrong-length genome")
	}
	if err := p.SetMemberGenome(99, g); err == nil {
		t.Fatal("expected error for out-of-range index")
	}
}

type fakeState []float64

func (s fakeState) Observation() []float64 { return s }

func collect(p *Population) []core.Individual {
	var out []core.Individual
	for m := range p.All() {
		out = append(out, m)
	}
	return out
}

func assignFitness(p *Population) {
	for i, m := range collect(p) {
		m.SetFitness(core.Reward(i))
	}
}

func containsGenome(p *Population, g []float64) bool {
	for m := range p.All() {
		if slices.Equal(m.Genome(), g) {
			return true
		}
	}
	return false
}
