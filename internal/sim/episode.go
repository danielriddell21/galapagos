package sim

import (
	"math/rand/v2"

	"github.com/danielriddell21/galapagos/internal/core"
)

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

func TrainAgent(env core.Environment, agent core.Agent, episodes, maxSteps int, seed int64) []core.Reward {
	rewards := make([]core.Reward, episodes)
	for i := range episodes {
		rewards[i] = RunEpisode(env, agent, maxSteps, TrackRNG(seed))
	}
	return rewards
}

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

func NewLiveEpisode(env core.Environment, agent core.Agent, maxSteps int, seed int64) *LiveEpisode {
	l := &LiveEpisode{env: env, agent: agent, maxSteps: maxSteps, seed: seed}
	l.reset()
	return l
}

func (l *LiveEpisode) reset() {
	l.state = l.env.Reset(TrackRNG(l.seed))
	l.step = 0
	l.ret = 0
	l.done = false
}

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

func (l *LiveEpisode) NextEpisode() {
	l.episode++
	l.reset()
}

func (l *LiveEpisode) Regenerate(seed int64) {
	l.seed = seed
	l.episode = 0
	l.reset()
}

func (l *LiveEpisode) Return() core.Reward { return l.ret }

func (l *LiveEpisode) Episode() int { return l.episode }

func (l *LiveEpisode) StepCount() int { return l.step }

func (l *LiveEpisode) Env() core.Environment { return l.env }
