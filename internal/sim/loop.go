package sim

import (
	"github.com/danielriddell21/galapagos/internal/core"
)

// collect gathers a population's members into an index-stable slice so they can
// be addressed positionally during a generation.
func collect(pop core.PopulationAgent) []core.Individual {
	members := make([]core.Individual, 0, pop.Len())
	for ind := range pop.All() {
		members = append(members, ind)
	}
	return members
}

// RunGeneration evaluates every member of pop simultaneously on the shared
// multi-environment for at most maxSteps, accumulating each member's reward as
// its fitness. The track is generated deterministically from seed. If onStep is
// non-nil it is called after each step (the render/pacing hook). The returned
// slice is indexed to match pop.All().
//
// Because bodies never interact (collisions are against walls only), this
// simultaneous rollout produces the same per-member fitness as evaluating each
// member independently; see EvaluateParallel.
func RunGeneration(env core.MultiEnvironment, pop core.PopulationAgent, maxSteps int, seed int64, onStep func()) []float64 {
	members := collect(pop)
	n := len(members)
	states := env.ResetAll(n, TrackRNG(seed))
	fitness := make([]float64, n)
	actions := make([]core.Action, n)
	dones := make([]bool, n)

	for range maxSteps {
		if env.Alive() == 0 {
			break
		}
		for i := range n {
			if dones[i] {
				actions[i] = nil
				continue
			}
			actions[i] = members[i].Act(states[i])
		}
		var rewards []core.Reward
		states, rewards, dones = env.StepAll(actions)
		for i := range n {
			fitness[i] += float64(rewards[i])
		}
		if onStep != nil {
			onStep()
		}
	}

	for i := range n {
		members[i].SetFitness(core.Reward(fitness[i]))
	}
	return fitness
}
