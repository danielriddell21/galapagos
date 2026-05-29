package sim

import (
	"github.com/danielriddell21/galapagos/internal/core"
)

// Live drives a population on a multi-environment one step at a time, so a
// renderer can advance the simulation frame by frame. Stepping a Live to the
// end of a generation yields the same per-member fitness as RunGeneration.
type Live struct {
	env      core.MultiEnvironment
	pop      core.PopulationAgent
	maxSteps int
	seed     int64

	members []core.Individual
	states  []core.State
	actions []core.Action
	dones   []bool
	fitness []float64
	step    int
}

// NewLive creates a live runner and starts the first generation.
func NewLive(env core.MultiEnvironment, pop core.PopulationAgent, maxSteps int, seed int64) *Live {
	l := &Live{env: env, pop: pop, maxSteps: maxSteps, seed: seed}
	l.startGeneration()
	return l
}

// startGeneration resets the environment and per-body bookkeeping for a fresh
// evaluation of the current population on the track for l.seed.
func (l *Live) startGeneration() {
	l.members = collect(l.pop)
	n := len(l.members)
	l.states = l.env.ResetAll(n, TrackRNG(l.seed))
	l.actions = make([]core.Action, n)
	l.dones = make([]bool, n)
	l.fitness = make([]float64, n)
	l.step = 0
}

// Step advances the simulation one tick and reports whether the generation has
// finished (all bodies done or the step budget reached). When it returns true,
// each member's fitness has been recorded.
func (l *Live) Step() (finished bool) {
	if l.finished() {
		return true
	}
	for i := range l.members {
		if l.dones[i] {
			l.actions[i] = nil
			continue
		}
		l.actions[i] = l.members[i].Act(l.states[i])
	}
	var rewards []core.Reward
	l.states, rewards, l.dones = l.env.StepAll(l.actions)
	for i := range l.members {
		l.fitness[i] += float64(rewards[i])
	}
	l.step++

	if l.finished() {
		for i := range l.members {
			l.members[i].SetFitness(core.Reward(l.fitness[i]))
		}
		return true
	}
	return false
}

func (l *Live) finished() bool {
	return l.env.Alive() == 0 || l.step >= l.maxSteps
}

// NextGeneration evolves the population and begins evaluating the new one.
func (l *Live) NextGeneration() {
	l.pop.Evolve()
	l.startGeneration()
}

// Regenerate switches to a new track seed and restarts the current generation
// on it, without evolving.
func (l *Live) Regenerate(seed int64) {
	l.seed = seed
	l.startGeneration()
}

// BestIndex returns the index of the current best member by accumulated fitness.
func (l *Live) BestIndex() int {
	best, bestFit := 0, l.fitness[0]
	for i, f := range l.fitness {
		if f > bestFit {
			best, bestFit = i, f
		}
	}
	return best
}

// Fitness returns the current accumulated fitness per member.
func (l *Live) Fitness() []float64 { return l.fitness }

// Step number within the current generation.
func (l *Live) StepCount() int { return l.step }
