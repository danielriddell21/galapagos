package main

import (
	"log/slog"
	"os"

	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/maze"
	"github.com/spf13/cobra"
)

func init() {
	var (
		agent       string
		size        int
		episodes    int
		generations int
		population  int
		maxSteps    int
		seed        int64
		headless    bool
	)
	cmd := &cobra.Command{
		Use:   "maze",
		Short: "Solve a grid maze with q-learning, ga, or neat",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			return runEnv(envParams{
				title:       "Galapagos — maze (" + agent + ")",
				newSingle:   func() core.Environment { return maze.New(maze.Config{Width: size, Height: size, MaxSteps: maxSteps}) },
				agent:       agent,
				seed:        resolveSeed(cmd, 0, log),
				headless:    headless,
				generations: generations,
				population:  population,
				episodes:    episodes,
				maxSteps:    maxSteps,
				qbins:       size,
			}, log)
		},
	}
	cmd.Flags().StringVar(&agent, "agent", "qlearning", "agent: qlearning, ga, or neat")
	cmd.Flags().IntVar(&size, "size", 6, "maze width and height in cells")
	cmd.Flags().IntVar(&episodes, "episodes", 3000, "episodes (qlearning)")
	cmd.Flags().IntVar(&generations, "generations", 30, "generations (ga/neat)")
	cmd.Flags().IntVar(&population, "population", 60, "population size (ga/neat)")
	cmd.Flags().IntVar(&maxSteps, "max-steps", 300, "maximum steps per episode")
	cmd.Flags().Int64Var(&seed, "seed", 0, "run seed (default: random, logged)")
	cmd.Flags().BoolVar(&headless, "headless", false, "train without a window")
	rootCmd.AddCommand(cmd)
}
