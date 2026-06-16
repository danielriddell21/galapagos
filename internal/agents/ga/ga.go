package ga

import (
	"fmt"
	"iter"
	"math/rand/v2"
	"slices"

	"github.com/danielriddell21/galapagos/internal/core"
)

// Distinct PCG stream constants keep weight initialization independent from
// reproduction, and both derive purely from the seed (and index/generation) so
// runs are reproducible regardless of evaluation order.
const (
	streamInit      uint64 = 0x53c5ff8e3d9b1a77
	streamEvolve    uint64 = 0x6a09e667f3bcc908
	streamImmigrant uint64 = 0x9e3779b97f4a7c15
)

// Config parameters the genetic algorithm. Inputs and Outputs are set from the
// environment's specs; the remaining fields come from the YAML config.
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

	// Diversity controls that resist premature convergence (the population
	// collapsing to clones of the elite and stalling in a local optimum). Both
	// default to sensible values in New when left zero.
	//
	// ImmigrantFraction is the share of each generation reseeded with fresh
	// random genomes, keeping a permanent floor of genetic diversity.
	//
	// When the best fitness has not improved for StagnationWindow generations the
	// breeding mutation rate and spread are scaled by HyperMutation, perturbing
	// the converged population hard enough to escape the optimum, then relaxing
	// once progress resumes.
	ImmigrantFraction float64
	StagnationWindow  int
	HyperMutation     float64
}

// vecAction is a generic action carrying a raw control vector, so the agent
// produces actions without importing any environment package.
type vecAction []float64

// Vector implements core.Action.
func (a vecAction) Vector() []float64 { return a }

// individual is one population member: a network plus its measured fitness.
type individual struct {
	genome  []float64
	net     *net
	fitness core.Reward
}

func newIndividual(cfg Config, genome []float64) *individual {
	return &individual{genome: genome, net: newNet(cfg.Inputs, cfg.HiddenSize, cfg.Outputs, genome)}
}

// Act returns the network's raw outputs as the action vector. Interpretation
// and clamping are left to the environment, which keeps the agent independent
// of any particular action semantics.
func (m *individual) Act(s core.State) core.Action {
	return vecAction(m.net.forward(s.Observation()))
}

func (m *individual) Fitness() core.Reward     { return m.fitness }
func (m *individual) SetFitness(r core.Reward) { m.fitness = r }
func (m *individual) Genome() []float64        { return m.genome }

// Population is a genetic-algorithm agent: a fixed-size set of networks evolved
// across generations. It implements core.PopulationAgent.
type Population struct {
	cfg     Config
	members []*individual
	gen     int

	// Stagnation tracking for adaptive mutation. bestSeen is the best fitness
	// observed so far; stalled counts generations since it last improved.
	bestSeen core.Reward
	stalled  int
	hasBest  bool
}

// New creates an initial population with Gaussian-random genomes. Each member's
// genome is seeded from the configured seed plus its index, so the starting
// population is reproducible.
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

// All yields the members in index order.
func (p *Population) All() iter.Seq[core.Individual] {
	return func(yield func(core.Individual) bool) {
		for _, m := range p.members {
			if !yield(m) {
				return
			}
		}
	}
}

// Len returns the population size.
func (p *Population) Len() int { return len(p.members) }

// Generation returns the current generation number, starting at zero.
func (p *Population) Generation() int { return p.gen }

// Act delegates to the current best member, so a Population can also be used as
// a plain agent (for replay or single-agent environments).
func (p *Population) Act(s core.State) core.Action { return p.best().Act(s) }

// Observe is unused: the GA learns from episode fitness, not per-step feedback.
func (p *Population) Observe(s core.State, a core.Action, r core.Reward, next core.State, done bool) {
}

// EndEpisode is unused: fitness is recorded on each member by the simulation.
func (p *Population) EndEpisode(total core.Reward) {}

// Best returns the genome of the fittest member, for saving.
func (p *Population) Best() []float64 { return clone(p.best().genome) }

// BestPolicy compiles the fittest member's network over a cloned genome into a
// standalone forward function, fixed against later evolution and safe to call
// concurrently. It backs hall-of-fame co-evolution, where a frozen champion is
// the opponent for the next generation.
func (p *Population) BestPolicy() func(obs []float64) []float64 {
	return newNet(p.cfg.Inputs, p.cfg.HiddenSize, p.cfg.Outputs, clone(p.best().genome)).forward
}

// SetMemberGenome replaces member i's genome, for injecting a loaded driver
// into a running population. The genome must match the network shape.
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

// best returns the fittest member.
func (p *Population) best() *individual {
	return slices.MaxFunc(p.members, func(a, b *individual) int {
		return cmpReward(a.fitness, b.fitness)
	})
}

// Evolve produces the next generation: elites are cloned unmutated, a slice of
// fresh random immigrants is injected to keep the gene pool diverse, and the
// remainder are bred by tournament selection, crossover, and mutation. When the
// best fitness has stalled, the breeding mutation is scaled up to break out of a
// local optimum. All randomness derives from the seed and generation, so
// reproduction is deterministic and independent of how fitness was evaluated.
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

// tournament samples size members and returns the fittest, biasing selection
// toward stronger genomes while preserving diversity.
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
