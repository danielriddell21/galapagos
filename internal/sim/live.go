package sim

import (
	"github.com/danielriddell21/galapagos/internal/core"
)

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

func NewLive(env core.MultiEnvironment, pop core.PopulationAgent, maxSteps int, seed int64) *Live {
	l := &Live{env: env, pop: pop, maxSteps: maxSteps, seed: seed}
	l.startGeneration()
	return l
}

func (l *Live) startGeneration() {
	l.members = collect(l.pop)
	n := len(l.members)
	l.states = l.env.ResetAll(n, TrackRNG(l.seed))
	l.actions = make([]core.Action, n)
	l.dones = make([]bool, n)
	l.fitness = make([]float64, n)
	l.step = 0
}

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

func (l *Live) NextGeneration() {
	l.pop.Evolve()
	l.startGeneration()
}

func (l *Live) Regenerate(seed int64) {
	l.seed = seed
	l.startGeneration()
}

func (l *Live) BestIndex() int {
	best, bestFit := 0, l.fitness[0]
	for i, f := range l.fitness {
		if f > bestFit {
			best, bestFit = i, f
		}
	}
	return best
}

func (l *Live) Fitness() []float64 { return l.fitness }

func (l *Live) StepCount() int { return l.step }
