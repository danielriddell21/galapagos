package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/danielriddell21/galapagos/internal/agents/ga"
	"github.com/danielriddell21/galapagos/internal/config"
	"github.com/danielriddell21/galapagos/internal/sim"
	"github.com/spf13/cobra"
)

func init() {
	var (
		cfgPath  string
		headless bool
		out      string
		seed     int64
	)
	cmd := &cobra.Command{
		Use:   "race",
		Short: "Evolve a population of cars on a procedural track",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.DefaultRacing()
			if cfgPath != "" {
				loaded, err := config.LoadRacing(cfgPath)
				if err != nil {
					return err
				}
				cfg = loaded
			}
			if cmd.Flags().Changed("seed") {
				cfg.Seed = seed
			}
			if headless {
				return runHeadless(cfg, out)
			}
			return launchGUI(cfg, out)
		},
	}
	cmd.Flags().StringVar(&cfgPath, "config", "", "path to a YAML config (defaults built in)")
	cmd.Flags().BoolVar(&headless, "headless", false, "train without a window, then save the best genome")
	cmd.Flags().StringVar(&out, "out", "best.json", "path to save the best genome")
	cmd.Flags().Int64Var(&seed, "seed", cfg().Seed, "override the config seed")
	rootCmd.AddCommand(cmd)
}

// cfg is a tiny helper so the flag default reflects the built-in seed.
func cfg() config.Racing { return config.DefaultRacing() }

// runHeadless evolves the population at full speed with no rendering and saves
// the best genome.
func runHeadless(c config.Racing, out string) error {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	pop := ga.New(gaConfigFrom(c))
	rc := racingConfigFrom(c)
	tel := sim.NewTelemetry(c.Generations, log)

	log.Info("training", "generations", c.Generations, "population", c.Population, "seed", c.Seed)
	sim.TrainHeadless(envFactory(rc), pop, sim.RunConfig{
		Seed:        c.Seed,
		MaxSteps:    c.MaxSteps,
		Generations: c.Generations,
	}, tel)

	if err := pop.SaveBest(out); err != nil {
		return err
	}
	fmt.Printf("saved best genome to %s\n", out)
	return nil
}
