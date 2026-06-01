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

// taskRNG returns the reset stream for a fixed-task episode at the given seed,
// matching the stream TrainAgent uses, so live and batch runs share layouts.
func taskRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewPCG(uint64(seed), streamTrack))
}

// LiveEpisode drives a single online agent on an environment one step at a time,
// so a renderer can show learning frame by frame. Stepping it to the end of an
// episode applies the same transitions as RunEpisode.
type LiveEpisode struct {
	env      core.Environment
	agent    core.Agent
	maxSteps int
	seed     int64

	state   core.State
	step    int
	episode int
	ret     core.Reward
	done    bool
}

// NewLiveEpisode creates a live single-agent runner and resets the first episode.
func NewLiveEpisode(env core.Environment, agent core.Agent, maxSteps int, seed int64) *LiveEpisode {
	l := &LiveEpisode{env: env, agent: agent, maxSteps: maxSteps, seed: seed}
	l.reset()
	return l
}

// reset begins an episode on the fixed task for the current seed.
func (l *LiveEpisode) reset() {
	l.state = l.env.Reset(taskRNG(l.seed))
	l.step = 0
	l.ret = 0
	l.done = false
}

// Step advances one transition, feeding the agent its feedback, and reports
// whether the episode has finished. When it returns true, the agent's
// EndEpisode hook has been called.
func (l *LiveEpisode) Step() (finished bool) {
	if l.done {
		return true
	}
	a := l.agent.Act(l.state)
	next, r, done := l.env.Step(a)
	l.agent.Observe(l.state, a, r, next, done)
	l.ret += r
	l.state = next
	l.step++
	if done || l.step >= l.maxSteps {
		l.done = true
		l.agent.EndEpisode(l.ret)
		return true
	}
	return false
}

// NextEpisode starts a fresh episode on the same task.
func (l *LiveEpisode) NextEpisode() {
	l.episode++
	l.reset()
}

// Regenerate switches to a new task seed and restarts the episode counter.
func (l *LiveEpisode) Regenerate(seed int64) {
	l.seed = seed
	l.episode = 0
	l.reset()
}

// Return reports the reward accumulated in the current episode.
func (l *LiveEpisode) Return() core.Reward { return l.ret }

// Episode reports the current episode number, starting at zero.
func (l *LiveEpisode) Episode() int { return l.episode }

// StepCount reports the step within the current episode.
func (l *LiveEpisode) StepCount() int { return l.step }

// Env exposes the environment for rendering.
func (l *LiveEpisode) Env() core.Environment { return l.env }
