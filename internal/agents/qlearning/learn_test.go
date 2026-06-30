package qlearning_test

import (
	"testing"

	"github.com/danielriddell21/galapagos/internal/agents/qlearning"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/cartpole"
	"github.com/danielriddell21/galapagos/internal/envs/maze"
	"github.com/danielriddell21/galapagos/internal/sim"
)

func TestBalancesCartpole(t *testing.T) {
	env := cartpole.New(cartpole.Config{MaxSteps: 500})
	agent := qlearning.New(qlearning.Config{
		Bins:         8,
		Actions:      2,
		Alpha:        0.1,
		Gamma:        0.99,
		Epsilon:      1.0,
		EpsilonDecay: 0.999,
		EpsilonMin:   0.01,
		Seed:         3,
	})

	rewards := sim.TrainAgent(env, agent, 4000, 500, 1)
	early := mean(rewards[:200])
	late := mean(rewards[len(rewards)-200:])
	if late <= early {
		t.Fatalf("no improvement: early=%.1f late=%.1f", early, late)
	}
	if late < 30 {
		t.Fatalf("late average balance %.1f too low; agent did not learn", late)
	}
}

func TestLearnsMaze(t *testing.T) {
	env := maze.New(maze.Config{Width: 6, Height: 6, MaxSteps: 300})
	agent := qlearning.New(qlearning.Config{
		Bins:         6,
		Actions:      4,
		Alpha:        0.3,
		Gamma:        0.99,
		Epsilon:      1.0,
		EpsilonDecay: 0.995,
		EpsilonMin:   0.01,
		Seed:         7,
	})

	rewards := sim.TrainAgent(env, agent, 4000, 300, 1)

	early := mean(rewards[:100])
	late := mean(rewards[len(rewards)-100:])
	if late <= early {
		t.Fatalf("no improvement: early=%.3f late=%.3f", early, late)
	}
	// A solved episode ends with the +1 goal reward dominating the small step
	// costs, so the late average should be clearly positive.
	if late < 0.5 {
		t.Fatalf("late average reward %.3f too low; agent did not learn to solve", late)
	}
}

func mean(xs []core.Reward) float64 {
	var s float64
	for _, x := range xs {
		s += float64(x)
	}
	return s / float64(len(xs))
}
