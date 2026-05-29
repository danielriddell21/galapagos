package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/danielriddell21/galapagos/internal/agents/qlearning"
	"github.com/danielriddell21/galapagos/internal/envs/maze"
	"github.com/danielriddell21/galapagos/internal/sim"
	"github.com/spf13/cobra"
)

func init() {
	var (
		episodes int
		size     int
		maxSteps int
		seed     int64
	)
	cmd := &cobra.Command{
		Use:   "maze",
		Short: "Solve a grid maze with tabular Q-learning",
		Long:  "Trains a tabular Q-learning agent on a generated maze, demonstrating a learning paradigm entirely different from the genetic algorithm on the same interfaces.",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			env := maze.New(maze.Config{Width: size, Height: size, MaxSteps: maxSteps})
			agent := qlearning.New(qlearning.Config{
				Bins: size, Actions: 4, Alpha: 0.3, Gamma: 0.99,
				Epsilon: 1.0, EpsilonDecay: 0.995, EpsilonMin: 0.01, Seed: seed,
			})

			rewards := sim.TrainAgent(env, agent, episodes, maxSteps, seed)
			log.Info("trained", "episodes", episodes, "states_seen", agent.States(), "final_epsilon", agent.Epsilon())
			final := rewards[len(rewards)-1]
			fmt.Printf("final episode reward %.3f (%s)\n", float64(final), solvedLabel(final > 0.5))
			return nil
		},
	}
	cmd.Flags().IntVar(&episodes, "episodes", 3000, "number of training episodes")
	cmd.Flags().IntVar(&size, "size", 6, "maze width and height in cells")
	cmd.Flags().IntVar(&maxSteps, "max-steps", 300, "maximum steps per episode")
	cmd.Flags().Int64Var(&seed, "seed", 1, "random seed")
	rootCmd.AddCommand(cmd)
}

func solvedLabel(solved bool) string {
	if solved {
		return "solved"
	}
	return "not solved"
}
