package sim

import (
	"runtime"
	"sync"

	"github.com/danielriddell21/galapagos/internal/core"
)

type EnvFactory func() core.MultiEnvironment

func EvaluateParallel(newEnv EnvFactory, members []core.Individual, maxSteps int, seed int64) []float64 {
	n := len(members)
	fitness := make([]float64, n)
	sem := make(chan struct{}, max(runtime.GOMAXPROCS(0), 1))
	var wg sync.WaitGroup

	for i := range n {
		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()
			fitness[i] = rolloutOne(newEnv(), members[i], maxSteps, seed)
		})
	}
	wg.Wait()

	for i := range n {
		members[i].SetFitness(core.Reward(fitness[i]))
	}
	return fitness
}

func rolloutOne(env core.MultiEnvironment, member core.Individual, maxSteps int, seed int64) float64 {
	states := env.ResetAll(1, TrackRNG(seed))
	actions := make([]core.Action, 1)
	var total float64
	done := false
	for range maxSteps {
		if done {
			break
		}
		actions[0] = member.Act(states[0])
		var rewards []core.Reward
		var dones []bool
		states, rewards, dones = env.StepAll(actions)
		total += float64(rewards[0])
		done = dones[0]
	}
	return total
}
