package neat

import (
	"iter"
	"math/rand/v2"
	"slices"

	"github.com/danielriddell21/galapagos/internal/core"
)

const (
	streamInit   uint64 = 0x1a2b3c4d5e6f7081
	streamEvolve uint64 = 0x0918273645546372
)

type Config struct {
	Population         int
	Inputs, Outputs    int
	WeightMutationRate float64
	WeightMutationStd  float64
	AddConnRate        float64
	AddNodeRate        float64
	CompatThreshold    float64
	C1, C2, C3         float64
	ElitePerSpecies    int
	SurvivalThreshold  float64
	Seed               int64
}

func DefaultConfig() Config {
	return Config{
		Population:         100,
		WeightMutationRate: 0.8,
		WeightMutationStd:  0.2,
		AddConnRate:        0.1,
		AddNodeRate:        0.05,
		CompatThreshold:    3.0,
		C1:                 1.0,
		C2:                 1.0,
		C3:                 0.4,
		ElitePerSpecies:    1,
		SurvivalThreshold:  0.5,
		Seed:               42,
	}
}

type Population struct {
	cfg     Config
	members []*genome
	inno    *innovations
	gen     int
}

func New(cfg Config) *Population {
	if cfg.ElitePerSpecies < 1 {
		cfg.ElitePerSpecies = 1
	}
	firstHidden := cfg.Inputs + 1 + cfg.Outputs
	inno := newInnovations(firstHidden)
	p := &Population{cfg: cfg, inno: inno, members: make([]*genome, cfg.Population)}
	for i := range cfg.Population {
		rng := rand.New(rand.NewPCG(uint64(cfg.Seed)^streamInit, uint64(i)))
		p.members[i] = newMinimalGenome(cfg.Inputs, cfg.Outputs, inno, rng)
	}
	inno.reset()
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

func (p *Population) BestPolicy() func(obs []float64) []float64 {
	return build(p.best().clone()).forward
}

func (p *Population) best() *genome {
	return slices.MaxFunc(p.members, func(a, b *genome) int {
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
	p.inno.reset()

	species := p.speciate()
	counts := p.allocate(species)

	var next []*genome
	for si, sp := range species {
		slices.SortFunc(sp, byFitnessDesc)
		quota := counts[si]
		if quota == 0 {
			continue
		}
		// Elitism: carry the best genomes over unmutated.
		elite := min(p.cfg.ElitePerSpecies, quota)
		for i := range elite {
			next = append(next, sp[i].clone())
		}
		parents := survivors(sp, p.cfg.SurvivalThreshold)
		for len(next) < cumulative(counts, si)+quota {
			next = append(next, p.breed(parents, rng))
		}
	}
	// Round-off can leave the population short or long; trim or top up.
	for len(next) > p.cfg.Population {
		next = next[:p.cfg.Population]
	}
	for len(next) < p.cfg.Population {
		next = append(next, p.breed(p.members, rng))
	}

	p.members = next
	p.gen++
}

func (p *Population) breed(pool []*genome, rng *rand.Rand) *genome {
	a := pool[rng.IntN(len(pool))]
	b := pool[rng.IntN(len(pool))]
	if b.fitness > a.fitness {
		a, b = b, a
	}
	child := crossover(a, b, rng)
	mutateWeights(child, p.cfg.WeightMutationRate, p.cfg.WeightMutationStd, rng)
	if rng.Float64() < p.cfg.AddConnRate {
		addConnection(child, p.inno, rng)
	}
	if rng.Float64() < p.cfg.AddNodeRate {
		addNode(child, p.inno, rng)
	}
	return child
}

func (p *Population) speciate() [][]*genome {
	var species [][]*genome
	for _, m := range p.members {
		placed := false
		for si := range species {
			rep := species[si][0]
			if distance(m, rep, p.cfg.C1, p.cfg.C2, p.cfg.C3) < p.cfg.CompatThreshold {
				species[si] = append(species[si], m)
				placed = true
				break
			}
		}
		if !placed {
			species = append(species, []*genome{m})
		}
	}
	return species
}

func (p *Population) allocate(species [][]*genome) []int {
	minFit := p.members[0].fitness
	for _, m := range p.members {
		minFit = min(minFit, m.fitness)
	}
	shared := make([]float64, len(species))
	var total float64
	for si, sp := range species {
		var sum float64
		for _, m := range sp {
			sum += float64(m.fitness-minFit) + 1 // shift to keep shares positive
		}
		shared[si] = sum / float64(len(sp)) // fitness sharing
		total += shared[si]
	}

	counts := make([]int, len(species))
	if total == 0 {
		return counts
	}
	type rem struct {
		idx  int
		frac float64
	}
	rems := make([]rem, 0, len(species))
	assigned := 0
	for si := range species {
		exact := shared[si] / total * float64(p.cfg.Population)
		counts[si] = int(exact)
		assigned += counts[si]
		rems = append(rems, rem{si, exact - float64(counts[si])})
	}
	slices.SortFunc(rems, func(a, b rem) int {
		switch {
		case a.frac > b.frac:
			return -1
		case a.frac < b.frac:
			return 1
		default:
			return a.idx - b.idx
		}
	})
	for i := 0; assigned < p.cfg.Population; i++ {
		counts[rems[i%len(rems)].idx]++
		assigned++
	}
	return counts
}

func byFitnessDesc(a, b *genome) int {
	switch {
	case a.fitness > b.fitness:
		return -1
	case a.fitness < b.fitness:
		return 1
	default:
		return 0
	}
}

func survivors(sorted []*genome, fraction float64) []*genome {
	n := max(1, int(float64(len(sorted))*fraction))
	return sorted[:n]
}

func cumulative(counts []int, i int) int {
	s := 0
	for j := range i {
		s += counts[j]
	}
	return s
}

var _ core.PopulationAgent = (*Population)(nil)
