package core

import (
	"iter"
	"math/rand/v2"
)

// Environment is a single-agent world. One agent acts on it per episode. The
// racing demo uses [MultiEnvironment] instead so a whole population can share
// one world; simple control tasks (cart-pole, maze) implement Environment.
type Environment interface {
	// Reset begins a new episode using rng for any procedural generation and
	// returns the initial state.
	Reset(rng *rand.Rand) State
	// Step applies an action and returns the next state, its reward, and
	// whether the episode has ended.
	Step(a Action) (State, Reward, bool)
	// ActionSpec describes the action vector the environment accepts.
	ActionSpec() Spec
	// ObservationSpec describes the observation vector the environment emits.
	ObservationSpec() Spec
	// Render draws the current world state through r.
	Render(r Renderer)
}

// MultiEnvironment is a world that advances many independent bodies on a single
// shared layout, so a population can be evaluated simultaneously. Bodies are
// addressed by a stable index: body i corresponds to actions[i] and states[i]
// for the life of an episode.
type MultiEnvironment interface {
	// ResetAll begins a new episode with n bodies and returns their initial
	// states. rng drives procedural layout generation.
	ResetAll(n int, rng *rand.Rand) []State
	// StepAll applies one action per body and returns the next states, rewards,
	// and done flags, each indexed to match the bodies.
	StepAll(actions []Action) (states []State, rewards []Reward, dones []bool)
	// ActionSpec describes the action vector each body accepts.
	ActionSpec() Spec
	// ObservationSpec describes the observation vector each body emits.
	ObservationSpec() Spec
	// Alive reports how many bodies have not yet finished.
	Alive() int
	// Render draws the shared world and all bodies through r.
	Render(r Renderer)
}

// Agent is a learning policy. It chooses actions, receives feedback, and is
// notified when an episode ends.
type Agent interface {
	// Act returns the action to take in state s.
	Act(s State) Action
	// Observe receives a transition for online learners. Episodic learners may
	// ignore it.
	Observe(s State, a Action, r Reward, next State, done bool)
	// EndEpisode reports the total reward accumulated over the finished episode.
	EndEpisode(totalReward Reward)
}

// Individual is one member of a population. It carries its own policy so the
// simulation can drive every member, and exposes fitness and genome through
// generic accessors so the renderer can rank and highlight members without
// knowing the concrete agent type.
type Individual interface {
	// Act returns the action this member's policy chooses in state s.
	Act(s State) Action
	// Fitness returns the member's most recent evaluated fitness.
	Fitness() Reward
	// SetFitness records the fitness measured for the member by the simulation.
	SetFitness(r Reward)
	// Genome returns the member's parameter vector.
	Genome() []float64
}

// PopulationAgent is the extended interface implemented by population-based
// learners such as the genetic algorithm and NEAT.
type PopulationAgent interface {
	Agent
	// All yields every member in stable index order.
	All() iter.Seq[Individual]
	// Len returns the population size.
	Len() int
	// Evolve produces the next generation from the current fitness values.
	Evolve()
	// Generation returns the current generation number, starting at zero.
	Generation() int
}
