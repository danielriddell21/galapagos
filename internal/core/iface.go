package core

import (
	"iter"
	"math/rand/v2"
)

type Environment interface {
	Reset(rng *rand.Rand) State

	Step(a Action) (State, Reward, bool)

	ActionSpec() Spec

	ObservationSpec() Spec

	Render(r Renderer)
}

type MultiEnvironment interface {
	ResetAll(n int, rng *rand.Rand) []State

	StepAll(actions []Action) (states []State, rewards []Reward, dones []bool)

	ActionSpec() Spec

	ObservationSpec() Spec

	Alive() int

	Render(r Renderer)
}

type Agent interface {
	Act(s State) Action

	Observe(s State, a Action, r Reward, next State, done bool)

	EndEpisode(totalReward Reward)
}

type Individual interface {
	Act(s State) Action

	Fitness() Reward

	SetFitness(r Reward)

	Genome() []float64
}

type PopulationAgent interface {
	Agent

	All() iter.Seq[Individual]

	Len() int

	Evolve()

	Generation() int
}
