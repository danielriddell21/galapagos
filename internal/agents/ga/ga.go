package ga

import (
	"fmt"
	"iter"
	"math/rand/v2"
	"slices"

	"github.com/danielriddell21/galapagos/internal/core"
	gacore "github.com/danielriddell21/galapagos/pkg/ga"
)

const streamInit uint64 = 0x53c5ff8e3d9b1a77

type Config struct {
	Population     int
	EliteFraction  float64
	MutationRate   float64
	MutationStd    float64
	HiddenSize     int
	Inputs         int
	Outputs        int
	TournamentSize int
	Seed           int64

	ImmigrantFraction float64
	StagnationWindow  int
	HyperMutation     float64
}

type vecAction []float64

func (a vecAction) Vector() []float64 { return a }

type individual struct {
	genome  []float64
	net     *net
	fitness core.Reward
}

func newIndividual(cfg Config, genome []float64) *individual {
	return &individual{genome: genome, net: newNet(cfg.Inputs, cfg.HiddenSize, cfg.Outputs, genome)}
}

func (m *individual) Act(s core.State) core.Action {
	return vecAction(m.net.forward(s.Observation()))
}

func (m *individual) Fitness() core.Reward     { return m.fitness }
func (m *individual) SetFitness(r core.Reward) { m.fitness = r }
func (m *individual) Genome() []float64        { return m.genome }

type Population struct {
	cfg     Config
	members []*individual
	gen     int

	bestSeen core.Reward
	stalled  int
	hasBest  bool
}

func New(cfg Config) *Population {
	if cfg.TournamentSize <= 0 {
		cfg.TournamentSize = 3
	}
	if cfg.ImmigrantFraction == 0 {
		cfg.ImmigrantFraction = 0.15
	}
	if cfg.StagnationWindow == 0 {
		cfg.StagnationWindow = 4
	}
	if cfg.HyperMutation == 0 {
		cfg.HyperMutation = 6
	}
	genomeLen := GenomeLen(cfg.Inputs, cfg.HiddenSize, cfg.Outputs)
	p := &Population{cfg: cfg, members: make([]*individual, cfg.Population)}
	for i := range cfg.Population {
		rng := rand.New(rand.NewPCG(uint64(cfg.Seed)^streamInit, uint64(i)))
		p.members[i] = newIndividual(cfg, gacore.RandomGenome(genomeLen, rng))
	}
	return p
}

func (p *Population) All() iter.Seq[core.Individual] {
	return func(yield func(core.Individual) bool) {
		for _, m := range p.members {
			if !yield(m) {
				return
			}
		}
	}
}

func (p *Population) Len() int { return len(p.members) }

func (p *Population) Generation() int { return p.gen }

func (p *Population) Act(s core.State) core.Action { return p.best().Act(s) }

func (p *Population) Observe(s core.State, a core.Action, r core.Reward, next core.State, done bool) {
}

func (p *Population) EndEpisode(total core.Reward) {}

func (p *Population) Best() []float64 { return slices.Clone(p.best().genome) }

func (p *Population) BestPolicy() func(obs []float64) []float64 {
	return newNet(p.cfg.Inputs, p.cfg.HiddenSize, p.cfg.Outputs, slices.Clone(p.best().genome)).forward
}

func (p *Population) SetMemberGenome(i int, genome []float64) error {
	if i < 0 || i >= len(p.members) {
		return fmt.Errorf("member index %d out of range [0,%d)", i, len(p.members))
	}
	if want := GenomeLen(p.cfg.Inputs, p.cfg.HiddenSize, p.cfg.Outputs); len(genome) != want {
		return fmt.Errorf("genome length %d does not match shape (want %d)", len(genome), want)
	}
	p.members[i] = newIndividual(p.cfg, slices.Clone(genome))
	return nil
}

func (p *Population) best() *individual {
	return slices.MaxFunc(p.members, func(a, b *individual) int {
		return cmpReward(a.fitness, b.fitness)
	})
}

func (p *Population) Evolve() {
	n := len(p.members)
	genomes := make([][]float64, n)
	fitness := make([]float64, n)
	for i, m := range p.members {
		genomes[i], fitness[i] = m.genome, float64(m.fitness)
	}

	// Track stagnation on the best fitness so the generic core can adapt mutation:
	// timid while the champion keeps improving, aggressive once it plateaus.
	best := core.Reward(fitness[0])
	for _, f := range fitness[1:] {
		best = max(best, core.Reward(f))
	}
	if !p.hasBest || best > p.bestSeen {
		p.bestSeen, p.stalled, p.hasBest = best, 0, true
	} else {
		p.stalled++
	}

	next := gacore.Reproduce(p.gaConfig(), genomes, fitness, p.gen, p.stalled)
	p.members = make([]*individual, n)
	for i, g := range next {
		p.members[i] = newIndividual(p.cfg, g)
	}
	p.gen++
}

func (p *Population) gaConfig() gacore.Config {
	return gacore.Config{
		EliteFraction:     p.cfg.EliteFraction,
		MutationRate:      p.cfg.MutationRate,
		MutationStd:       p.cfg.MutationStd,
		TournamentSize:    p.cfg.TournamentSize,
		ImmigrantFraction: p.cfg.ImmigrantFraction,
		StagnationWindow:  p.cfg.StagnationWindow,
		HyperMutation:     p.cfg.HyperMutation,
		Seed:              p.cfg.Seed,
	}
}

func cmpReward(a, b core.Reward) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

var _ core.PopulationAgent = (*Population)(nil)
