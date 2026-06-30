package cube_test

import (
	"testing"

	"github.com/danielriddell21/galapagos/internal/agents/qlearning"
	"github.com/danielriddell21/galapagos/internal/envs/cube"
	"github.com/danielriddell21/galapagos/internal/sim"
)

func TestQLearningSolvesShallowScramble(t *testing.T) {
	env := cube.New(cube.Config{ScrambleDepth: 3, MaxSteps: 8})
	agent := qlearning.New(qlearning.Config{
		Actions: 18, Alpha: 0.3, Gamma: 0.99,
		Epsilon: 1.0, EpsilonDecay: 0.995, EpsilonMin: 0.01,
		Seed: 7, KeyByState: true,
	})

	rewards := sim.TrainAgent(env, agent, 8000, 8, 7)
	if final := rewards[len(rewards)-1]; final <= 0 {
		t.Fatalf("did not learn to solve: final episode return %.3f", float64(final))
	}
}
