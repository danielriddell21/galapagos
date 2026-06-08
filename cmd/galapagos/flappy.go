package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/flappy"
	"github.com/spf13/cobra"
)

func init() {
	var (
		agent       string
		generations int
		population  int
		maxSteps    int
		seed        int64
		headless    bool
	)
	cmd := &cobra.Command{
		Use:   "flappy",
		Short: "Evolve a swarm of birds to flap through scrolling pipe gaps",
		Long: "Evolves a population (ga or neat) to fly birds through scrolling pipe gaps. " +
			"The window shows the whole swarm on one shared course, each bird flapping by its " +
			"own evolved network and crashing out until the fittest survive.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if agent != "ga" && agent != "neat" {
				return fmt.Errorf("flappy supports only the ga and neat agents")
			}
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			return runEnv(envParams{
				title:       "Galapagos — flappy (" + agent + ")",
				newSingle:   func() core.Environment { return flappy.New(flappy.Config{MaxSteps: maxSteps}) },
				agent:       agent,
				seed:        resolveSeed(cmd, 0, log),
				headless:    headless,
				generations: generations,
				population:  population,
				maxSteps:    maxSteps,
			}, log)
		},
	}
	cmd.Flags().StringVar(&agent, "agent", "ga", "agent: ga or neat")
	cmd.Flags().IntVar(&generations, "generations", 40, "generations to evolve")
	cmd.Flags().IntVar(&population, "population", 80, "population size (birds in the swarm)")
	cmd.Flags().IntVar(&maxSteps, "max-steps", 600, "maximum steps per episode")
	cmd.Flags().Int64Var(&seed, "seed", 0, "run seed (default: random, logged)")
	cmd.Flags().BoolVar(&headless, "headless", false, "train without a window")
	rootCmd.AddCommand(cmd)
}
