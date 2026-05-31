package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/cube"
	"github.com/spf13/cobra"
)

func init() {
	var (
		agent    string
		scramble int
		episodes int
		maxSteps int
		seed     int64
		headless bool
	)
	cmd := &cobra.Command{
		Use:   "cube",
		Short: "Solve a scrambled Rubik's cube with tabular Q-learning",
		Long:  "Trains tabular Q-learning on a Rubik's cube, keying its Q-table directly on the comparable cube state. It learns to solve one reproducible scramble per run.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if agent != "qlearning" {
				return fmt.Errorf("cube supports only the qlearning agent (the cube state space defeats a single-output ga/neat policy)")
			}
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			return runEnv(envParams{
				title:     "Galapagos — cube (qlearning)",
				newSingle: func() core.Environment { return cube.New(cube.Config{ScrambleDepth: scramble, MaxSteps: maxSteps}) },
				agent:     agent,
				seed:      resolveSeed(cmd, 0, log),
				headless:  headless,
				episodes:  episodes,
				maxSteps:  maxSteps,
				qByState:  true,
			}, log)
		},
	}
	cmd.Flags().StringVar(&agent, "agent", "qlearning", "agent (cube supports qlearning only)")
	// Tabular Q-learning needs a tight horizon so exploration stays near the goal;
	// shallow scrambles solve reliably, deeper ones need far more episodes.
	cmd.Flags().IntVar(&scramble, "scramble", 4, "number of moves used to scramble the start")
	cmd.Flags().IntVar(&episodes, "episodes", 20000, "training episodes")
	cmd.Flags().IntVar(&maxSteps, "max-steps", 10, "maximum moves per episode")
	cmd.Flags().Int64Var(&seed, "seed", 0, "run seed (default: random, logged)")
	cmd.Flags().BoolVar(&headless, "headless", false, "train without a window")
	rootCmd.AddCommand(cmd)
}
