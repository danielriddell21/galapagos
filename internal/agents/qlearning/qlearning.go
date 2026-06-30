package qlearning

import (
	"math/rand/v2"
	"slices"

	"github.com/danielriddell21/galapagos/internal/core"
)

type Config struct {
	Bins         int
	Actions      int
	Alpha        float64
	Gamma        float64
	Epsilon      float64
	EpsilonDecay float64
	EpsilonMin   float64
	Seed         int64

	KeyByState bool
}

func DefaultConfig() Config {
	return Config{
		Bins:         8,
		Actions:      4,
		Alpha:        0.2,
		Gamma:        0.99,
		Epsilon:      1.0,
		EpsilonDecay: 0.99,
		EpsilonMin:   0.02,
		Seed:         42,
	}
}

type vecAction float64

func (a vecAction) Vector() []float64 { return []float64{float64(a)} }

type Agent struct {
	cfg     Config
	q       map[any][]float64
	rng     *rand.Rand
	epsilon float64
}

func New(cfg Config) *Agent {
	return &Agent{
		cfg:     cfg,
		q:       map[any][]float64{},
		rng:     rand.New(rand.NewPCG(uint64(cfg.Seed), 0xa1b2c3d4)),
		epsilon: cfg.Epsilon,
	}
}

func (a *Agent) keyOf(s core.State) any {
	if a.cfg.KeyByState {
		return s
	}
	return a.key(s.Observation())
}

func (a *Agent) key(obs []float64) int {
	k := 0
	for _, v := range obs {
		b := int(min(max(v, 0), 1) * float64(a.cfg.Bins))
		b = min(b, a.cfg.Bins-1)
		k = k*a.cfg.Bins + b
	}
	return k
}

func (a *Agent) values(k any) []float64 {
	row, ok := a.q[k]
	if !ok {
		row = make([]float64, a.cfg.Actions)
		a.q[k] = row
	}
	return row
}

func (a *Agent) Act(s core.State) core.Action {
	if a.rng.Float64() < a.epsilon {
		return vecAction(a.rng.IntN(a.cfg.Actions))
	}
	return vecAction(argmax(a.values(a.keyOf(s))))
}

func (a *Agent) Observe(s core.State, act core.Action, r core.Reward, next core.State, done bool) {
	idx := int(act.Vector()[0] + 0.5)
	row := a.values(a.keyOf(s))
	target := float64(r)
	if !done {
		target += a.cfg.Gamma * slices.Max(a.values(a.keyOf(next)))
	}
	row[idx] += a.cfg.Alpha * (target - row[idx])
}

func (a *Agent) EndEpisode(total core.Reward) {
	a.epsilon = max(a.cfg.EpsilonMin, a.epsilon*a.cfg.EpsilonDecay)
}

func (a *Agent) Epsilon() float64 { return a.epsilon }

func (a *Agent) States() int { return len(a.q) }

func argmax(v []float64) int {
	best := 0
	for i, x := range v {
		if x > v[best] {
			best = i
		}
	}
	return best
}

var _ core.Agent = (*Agent)(nil)
