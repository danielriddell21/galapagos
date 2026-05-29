package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/danielriddell21/galapagos/internal/agents/ga"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/cartpole"
	"github.com/danielriddell21/galapagos/internal/sim"
	"github.com/spf13/cobra"
)

func init() {
	var (
		generations int
		population  int
		maxSteps    int
		seed        int64
	)
	cmd := &cobra.Command{
		Use:   "cartpole",
		Short: "Evolve a cart-pole balancer with the genetic algorithm",
		Long:  "Runs the same genetic algorithm used for racing on the cart-pole task via the single-to-multi adapter, demonstrating the agent and environment interfaces are decoupled.",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			newEnv := func() core.MultiEnvironment {
				return sim.AsMulti(func() core.Environment { return cartpole.New(cartpole.Config{MaxSteps: maxSteps}) })
			}
			pop := ga.New(ga.Config{
				Population: population, EliteFraction: 0.1, MutationRate: 0.1, MutationStd: 0.3,
				HiddenSize: 6, Inputs: 4, Outputs: 1, Seed: seed,
			})
			tel := sim.NewTelemetry(generations, log)
			sim.TrainHeadless(newEnv, pop, sim.RunConfig{Seed: seed, MaxSteps: maxSteps, Generations: generations}, tel)

			best := sim.EvaluateParallel(newEnv, individuals(pop), maxSteps, seed)
			fmt.Printf("best balanced for %.0f of %d steps\n", maxOf(best), maxSteps)
			return nil
		},
	}
	cmd.Flags().IntVar(&generations, "generations", 30, "number of generations")
	cmd.Flags().IntVar(&population, "population", 60, "population size")
	cmd.Flags().IntVar(&maxSteps, "max-steps", 500, "maximum steps per episode")
	cmd.Flags().Int64Var(&seed, "seed", 42, "random seed")
	rootCmd.AddCommand(cmd)
}

// individuals collects a population's members for evaluation.
func individuals(pop *ga.Population) []core.Individual {
	var out []core.Individual
	for m := range pop.All() {
		out = append(out, m)
	}
	return out
}

func maxOf(xs []float64) float64 {
	m := xs[0]
	for _, x := range xs {
		if x > m {
			m = x
		}
	}
	return m
}
