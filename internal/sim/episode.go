package sim

import (
	"math/rand/v2"

	"github.com/danielriddell21/galapagos/internal/core"
)

// RunEpisode runs one episode of a single-agent environment, feeding the agent
// each transition, and returns the total reward. rng seeds the environment's
// reset.
func RunEpisode(env core.Environment, agent core.Agent, maxSteps int, rng *rand.Rand) core.Reward {
	s := env.Reset(rng)
	var total core.Reward
	for range maxSteps {
		a := agent.Act(s)
		next, r, done := env.Step(a)
		agent.Observe(s, a, r, next, done)
		total += r
		s = next
		if done {
			break
		}
	}
	agent.EndEpisode(total)
	return total
}

// TrainAgent trains an online agent over episodes on a fixed task: every episode
// resets the environment from the same derived stream, so the layout is
// identical and learning is reproducible. It returns the per-episode rewards.
func TrainAgent(env core.Environment, agent core.Agent, episodes, maxSteps int, seed int64) []core.Reward {
	rewards := make([]core.Reward, episodes)
	for i := range episodes {
		resetRNG := rand.New(rand.NewPCG(uint64(seed), streamTrack))
		rewards[i] = RunEpisode(env, agent, maxSteps, resetRNG)
	}
	return rewards
}
