package evochess

import (
	"iter"
	"math/rand/v2"
	"slices"

	gambit "github.com/danielriddell21/gambit/pkg/chess"

	"github.com/danielriddell21/galapagos/internal/core"
)

const (
	streamInit   uint64 = 0x9e3779b97f4a7c15
	streamEvolve uint64 = 0xc2b2ae3d27d4eb4f
)

type Config struct {
	Population     int
	Depth          int
	EliteFraction  float64
	MutationRate   float64
	MutationStd    float64
	TournamentSize int
	Seed           int64
}

func DefaultConfig() Config {
	return Config{
		Population: 24, Depth: 2, EliteFraction: 0.1,
		MutationRate: 0.1, MutationStd: 0.3, TournamentSize: 3,
	}
}

type boarder interface {
	Board() *gambit.Board
}

type individual struct {
	genome  []float64
	eval    *evaluator
	depth   int
	fitness core.Reward
}

func newIndividual(genome []float64, depth int) *individual {
	return &individual{genome: genome, eval: newEvaluator(genome), depth: depth}
}

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

type Population struct {
	cfg     Config
	members []*individual
	gen     int
}

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

func (p *Population) Observe(core.State, core.Action, core.Reward, core.State, bool) {}

func (p *Population) EndEpisode(core.Reward) {}

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

func tournament(pop []*individual, size int, rng *rand.Rand) *individual {
	best := pop[rng.IntN(len(pop))]
	for range size - 1 {
		if c := pop[rng.IntN(len(pop))]; c.fitness > best.fitness {
			best = c
		}
	}
	return best
}

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

func mutate(g []float64, rate, std float64, rng *rand.Rand) {
	for i := range g {
		if rng.Float64() < rate {
			g[i] += std * rng.NormFloat64()
		}
	}
}

var _ core.PopulationAgent = (*Population)(nil)
