package evochess

import (
	"iter"
	"math/rand/v2"
	"slices"

	gambit "github.com/danielriddell21/gambit/pkg/chess"

	"github.com/danielriddell21/galapagos/internal/core"
)

// PCG stream constants keep initialization and reproduction independent and
// reproducible from the seed.
const (
	streamInit   uint64 = 0x9e3779b97f4a7c15
	streamEvolve uint64 = 0xc2b2ae3d27d4eb4f
)

// Config parameters the evolving search agent.
type Config struct {
	Population     int
	Depth          int     // alpha-beta search depth in plies
	EliteFraction  float64 // top fraction carried over unmutated
	MutationRate   float64 // per-gene mutation probability
	MutationStd    float64 // mutation step size
	TournamentSize int
	Seed           int64
}

// DefaultConfig returns balanced parameters for a small, watchable run.
func DefaultConfig() Config {
	return Config{
		Population: 24, Depth: 2, EliteFraction: 0.1,
		MutationRate: 0.1, MutationStd: 0.3, TournamentSize: 3,
	}
}

// boarder is satisfied by the chess environment's state, exposing the position
// so the agent can search. The agent never imports the environment.
type boarder interface {
	Board() *gambit.Board
}

// individual is one member: an evolved evaluator plus its measured fitness.
type individual struct {
	genome  []float64
	eval    *evaluator
	depth   int
	fitness core.Reward
}

func newIndividual(genome []float64, depth int) *individual {
	return &individual{genome: genome, eval: newEvaluator(genome), depth: depth}
}

// Act searches the current position and returns the chosen move encoded for the
// chess environment. If the state is not a chess position it yields an empty
// action (the environment then falls back to a legal move).
func (m *individual) Act(s core.State) core.Action {
	b, ok := s.(boarder)
	if !ok {
		return action(make([]float64, 128))
	}
	return encodeMove(b.Board(), bestMove(m.eval, b.Board(), m.depth))
}

func (m *individual) Fitness() core.Reward     { return m.fitness }
func (m *individual) SetFitness(r core.Reward) { m.fitness = r }
func (m *individual) Genome() []float64        { return m.genome }

// Population is the evolving search agent. It implements core.PopulationAgent.
type Population struct {
	cfg     Config
	members []*individual
	gen     int
}

// New creates an initial population whose evaluators are seeded near classical
// material values with small random piece-square perturbations.
func New(cfg Config) *Population {
	if cfg.TournamentSize <= 0 {
		cfg.TournamentSize = 3
	}
	if cfg.Depth <= 0 {
		cfg.Depth = 2
	}
	p := &Population{cfg: cfg, members: make([]*individual, cfg.Population)}
	for i := range cfg.Population {
		rng := rand.New(rand.NewPCG(uint64(cfg.Seed)^streamInit, uint64(i)))
		p.members[i] = newIndividual(randomGenome(rng), cfg.Depth)
	}
	return p
}

// randomGenome seeds material near classical values and the piece-square tables
// with small Gaussian noise.
func randomGenome(rng *rand.Rand) []float64 {
	g := make([]float64, genomeLen)
	for t := range numTypes {
		g[t] = baseMaterial[t] + 0.5*rng.NormFloat64()
	}
	for i := numTypes; i < genomeLen; i++ {
		g[i] = 0.3 * rng.NormFloat64()
	}
	return g
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

// Generation returns the current generation number.
func (p *Population) Generation() int { return p.gen }

// Act delegates to the fittest member, so the population can act as a plain agent.
func (p *Population) Act(s core.State) core.Action { return p.best().Act(s) }

// Observe is unused: the agent learns from episode fitness.
func (p *Population) Observe(core.State, core.Action, core.Reward, core.State, bool) {}

// EndEpisode is unused: fitness is recorded per member by the simulation.
func (p *Population) EndEpisode(core.Reward) {}

// FrozenPolicy returns the current best member's move policy over a cloned
// evaluator, fixed against later evolution and safe to call concurrently. It
// backs hall-of-fame co-evolution.
func (p *Population) FrozenPolicy() func(core.State) core.Action {
	eval := p.best().eval.clone()
	depth := p.cfg.Depth
	return func(s core.State) core.Action {
		b, ok := s.(boarder)
		if !ok {
			return action(make([]float64, 128))
		}
		return encodeMove(b.Board(), bestMove(eval, b.Board(), depth))
	}
}

// best returns the fittest member.
func (p *Population) best() *individual {
	return slices.MaxFunc(p.members, func(a, b *individual) int {
		switch {
		case a.fitness < b.fitness:
			return -1
		case a.fitness > b.fitness:
			return 1
		default:
			return 0
		}
	})
}

// Evolve produces the next generation: elites carry over unmutated, the rest are
// bred by tournament selection, uniform crossover, and Gaussian mutation. All
// randomness derives from the seed and generation.
func (p *Population) Evolve() {
	rng := rand.New(rand.NewPCG(uint64(p.cfg.Seed)^streamEvolve, uint64(p.gen)))

	ranked := slices.Clone(p.members)
	slices.SortFunc(ranked, func(a, b *individual) int {
		switch {
		case a.fitness > b.fitness:
			return -1
		case a.fitness < b.fitness:
			return 1
		default:
			return 0
		}
	})

	eliteCount := max(1, int(p.cfg.EliteFraction*float64(len(p.members))))
	next := make([]*individual, 0, len(p.members))
	for i := range eliteCount {
		next = append(next, newIndividual(slices.Clone(ranked[i].genome), p.cfg.Depth))
	}
	for len(next) < len(p.members) {
		a := tournament(ranked, p.cfg.TournamentSize, rng)
		b := tournament(ranked, p.cfg.TournamentSize, rng)
		child := crossover(a.genome, b.genome, rng)
		mutate(child, p.cfg.MutationRate, p.cfg.MutationStd, rng)
		next = append(next, newIndividual(child, p.cfg.Depth))
	}

	p.members = next
	p.gen++
}

// tournament samples size members and returns the fittest.
func tournament(pop []*individual, size int, rng *rand.Rand) *individual {
	best := pop[rng.IntN(len(pop))]
	for range size - 1 {
		if c := pop[rng.IntN(len(pop))]; c.fitness > best.fitness {
			best = c
		}
	}
	return best
}

// crossover blends two genomes gene-by-gene with a uniform mask.
func crossover(a, b []float64, rng *rand.Rand) []float64 {
	child := make([]float64, len(a))
	for i := range child {
		if rng.IntN(2) == 0 {
			child[i] = a[i]
		} else {
			child[i] = b[i]
		}
	}
	return child
}

// mutate perturbs each gene with probability rate by Gaussian noise scaled by std.
func mutate(g []float64, rate, std float64, rng *rand.Rand) {
	for i := range g {
		if rng.Float64() < rate {
			g[i] += std * rng.NormFloat64()
		}
	}
}

var _ core.PopulationAgent = (*Population)(nil)
