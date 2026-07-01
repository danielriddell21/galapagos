package ga

import (
	"fmt"
	"iter"
	"math/rand/v2"
	"slices"

	"github.com/danielriddell21/galapagos/internal/core"
)

const (
	streamInit      uint64 = 0x53c5ff8e3d9b1a77
	streamEvolve    uint64 = 0x6a09e667f3bcc908
	streamImmigrant uint64 = 0x9e3779b97f4a7c15
)

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
		p.members[i] = newIndividual(cfg, randomGenome(genomeLen, rng))
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

func (p *Population) Best() []float64 { return clone(p.best().genome) }

func (p *Population) BestPolicy() func(obs []float64) []float64 {
	return newNet(p.cfg.Inputs, p.cfg.HiddenSize, p.cfg.Outputs, clone(p.best().genome)).forward
}

func (p *Population) SetMemberGenome(i int, genome []float64) error {
	if i < 0 || i >= len(p.members) {
		return fmt.Errorf("member index %d out of range [0,%d)", i, len(p.members))
	}
	if want := GenomeLen(p.cfg.Inputs, p.cfg.HiddenSize, p.cfg.Outputs); len(genome) != want {
		return fmt.Errorf("genome length %d does not match shape (want %d)", len(genome), want)
	}
	p.members[i] = newIndividual(p.cfg, clone(genome))
	return nil
}

func (p *Population) best() *individual {
	return slices.MaxFunc(p.members, func(a, b *individual) int {
		return cmpReward(a.fitness, b.fitness)
	})
}

func (p *Population) Evolve() {
	rng := rand.New(rand.NewPCG(uint64(p.cfg.Seed)^streamEvolve, uint64(p.gen)))
	n := len(p.members)

	ranked := slices.Clone(p.members)
	slices.SortFunc(ranked, func(a, b *individual) int {
		return cmpReward(b.fitness, a.fitness) // descending
	})

	// Track stagnation on the best fitness so mutation can adapt: timid while the
	// champion keeps improving, aggressive once it plateaus.
	if best := ranked[0].fitness; !p.hasBest || best > p.bestSeen {
		p.bestSeen, p.stalled, p.hasBest = best, 0, true
	} else {
		p.stalled++
	}
	rate, std := p.cfg.MutationRate, p.cfg.MutationStd
	if p.cfg.StagnationWindow > 0 && p.stalled >= p.cfg.StagnationWindow {
		// While the champion is stuck, widen the breeding mutation to explore past
		// the local optimum. A fixed, moderate boost works better than escalating
		// it further, which degrades into an unproductive random search.
		rate = min(1, rate*p.cfg.HyperMutation)
		std *= p.cfg.HyperMutation
	}

	eliteCount := max(1, int(p.cfg.EliteFraction*float64(n)))
	immigrantCount := int(p.cfg.ImmigrantFraction * float64(n))
	if eliteCount+immigrantCount > n {
		immigrantCount = n - eliteCount
	}

	next := make([]*individual, 0, n)
	for i := range eliteCount {
		next = append(next, newIndividual(p.cfg, clone(ranked[i].genome)))
	}
	genomeLen := GenomeLen(p.cfg.Inputs, p.cfg.HiddenSize, p.cfg.Outputs)
	for i := range immigrantCount {
		irng := rand.New(rand.NewPCG(uint64(p.cfg.Seed)^streamImmigrant, uint64(p.gen)*uint64(n)+uint64(i)))
		next = append(next, newIndividual(p.cfg, randomGenome(genomeLen, irng)))
	}
	for len(next) < n {
		a := tournament(ranked, p.cfg.TournamentSize, rng)
		b := tournament(ranked, p.cfg.TournamentSize, rng)
		child := crossover(a.genome, b.genome, rng)
		mutate(child, rate, std, rng)
		next = append(next, newIndividual(p.cfg, child))
	}

	p.members = next
	p.gen++
}

func tournament(pop []*individual, size int, rng *rand.Rand) *individual {
	best := pop[rng.IntN(len(pop))]
	for range size - 1 {
		c := pop[rng.IntN(len(pop))]
		if c.fitness > best.fitness {
			best = c
		}
	}
	return best
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
