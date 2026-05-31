// Package qlearning implements a tabular Q-learning agent. It discretizes the
// (already normalized) observation vector into bins and learns action values
// online. It depends only on core, proving the Agent interface generalizes
// across learning paradigms.
package qlearning

import (
	"math/rand/v2"
	"slices"

	"github.com/danielriddell21/galapagos/internal/core"
)

// Config parameters the agent. Actions is the number of discrete actions the
// environment accepts; the agent emits the chosen index as the action vector.
type Config struct {
	Bins         int     // bins per observation dimension (binned-observation keying)
	Actions      int     // number of discrete actions
	Alpha        float64 // learning rate
	Gamma        float64 // discount factor
	Epsilon      float64 // initial exploration rate
	EpsilonDecay float64 // multiplier applied to epsilon each episode
	EpsilonMin   float64 // floor for epsilon
	Seed         int64
	// KeyByState keys the Q-table on the State value itself instead of a binned
	// observation. The State's concrete type must be comparable (e.g. a cube
	// position); slice-backed observations (maze, cart-pole) must leave this false.
	KeyByState bool
}

// DefaultConfig returns sensible defaults for small discrete tasks.
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

// vecAction carries a discrete action index as a one-element vector.
type vecAction float64

// Vector implements core.Action.
func (a vecAction) Vector() []float64 { return []float64{float64(a)} }

// Agent is a tabular Q-learning learner.
type Agent struct {
	cfg     Config
	q       map[any][]float64
	rng     *rand.Rand
	epsilon float64
}

// New returns a Q-learning agent with an empty table.
func New(cfg Config) *Agent {
	return &Agent{
		cfg:     cfg,
		q:       map[any][]float64{},
		rng:     rand.New(rand.NewPCG(uint64(cfg.Seed), 0xa1b2c3d4)),
		epsilon: cfg.Epsilon,
	}
}

// keyOf returns the Q-table key for a state: the comparable state value itself
// when KeyByState is set, otherwise the binned-observation index.
func (a *Agent) keyOf(s core.State) any {
	if a.cfg.KeyByState {
		return s
	}
	return a.key(s.Observation())
}

// key encodes a discretized observation into a single table index using a
// mixed-radix scheme.
func (a *Agent) key(obs []float64) int {
	k := 0
	for _, v := range obs {
		b := int(min(max(v, 0), 1) * float64(a.cfg.Bins))
		b = min(b, a.cfg.Bins-1)
		k = k*a.cfg.Bins + b
	}
	return k
}

// values returns the action-value row for a state key, creating it if absent.
func (a *Agent) values(k any) []float64 {
	row, ok := a.q[k]
	if !ok {
		row = make([]float64, a.cfg.Actions)
		a.q[k] = row
	}
	return row
}

// Act selects an action with an epsilon-greedy policy.
func (a *Agent) Act(s core.State) core.Action {
	if a.rng.Float64() < a.epsilon {
		return vecAction(a.rng.IntN(a.cfg.Actions))
	}
	return vecAction(argmax(a.values(a.keyOf(s))))
}

// Observe applies the temporal-difference update for the taken transition.
func (a *Agent) Observe(s core.State, act core.Action, r core.Reward, next core.State, done bool) {
	idx := int(act.Vector()[0] + 0.5)
	row := a.values(a.keyOf(s))
	target := float64(r)
	if !done {
		target += a.cfg.Gamma * slices.Max(a.values(a.keyOf(next)))
	}
	row[idx] += a.cfg.Alpha * (target - row[idx])
}

// EndEpisode decays the exploration rate toward its floor.
func (a *Agent) EndEpisode(total core.Reward) {
	a.epsilon = max(a.cfg.EpsilonMin, a.epsilon*a.cfg.EpsilonDecay)
}

// Epsilon returns the current exploration rate, for inspection and logging.
func (a *Agent) Epsilon() float64 { return a.epsilon }

// States returns the number of distinct states seen, for diagnostics.
func (a *Agent) States() int { return len(a.q) }

// argmax returns the index of the largest value; ties resolve to the lowest
// index, keeping action selection deterministic.
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
