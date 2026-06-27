package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/cartpole"
)

func init() {
	var (
		agent       string
		generations int
		population  int
		episodes    int
		maxSteps    int
		seed        int64
		headless    bool
	)
	cmd := &cobra.Command{
		Use:   "cartpole",
		Short: "Balance a cart-pole with ga, neat, or q-learning",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			return runEnv(envParams{
				title:       "Galapagos — cartpole (" + agent + ")",
				newSingle:   func() core.Environment { return cartpole.New(cartpole.Config{MaxSteps: maxSteps}) },
				agent:       agent,
				seed:        resolveSeed(cmd, 0, log),
				headless:    headless,
				generations: generations,
				population:  population,
				episodes:    episodes,
				maxSteps:    maxSteps,
				qbins:       8,
			}, log)
		},
	}
	cmd.Flags().StringVar(&agent, "agent", "ga", "agent: ga, neat, or qlearning")
	cmd.Flags().IntVar(&generations, "generations", 30, "generations (ga/neat)")
	cmd.Flags().IntVar(&population, "population", 60, "population size (ga/neat)")
	cmd.Flags().IntVar(&episodes, "episodes", 4000, "episodes (qlearning)")
	cmd.Flags().IntVar(&maxSteps, "max-steps", 500, "maximum steps per episode")
	cmd.Flags().Int64Var(&seed, "seed", 0, "run seed (default: random, logged)")
	cmd.Flags().BoolVar(&headless, "headless", false, "train without a window")
	rootCmd.AddCommand(cmd)
}
