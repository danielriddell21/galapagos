package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/flappy"
)

// defaultFlappy is the flappy command's flag defaults. The demo clip runs with
// them unchanged, so the recorded media shows what the command does.
func defaultFlappy() envParams {
	return envParams{agent: "ga", generations: 40, population: 80, maxSteps: 2000}
}

// flappyParams fills in the parts of p that describe the environment itself,
// once the flags (or a demo clip) have settled the tunables.
func flappyParams(p *envParams) {
	p.title = "Galapagos — flappy (" + p.agent + ")"
	p.newSingle = func() core.Environment { return flappy.New(flappy.Config{MaxSteps: p.maxSteps}) }
}

func init() {
	p := defaultFlappy()
	var seed int64
	cmd := &cobra.Command{
		Use:   "flappy",
		Short: "Evolve a swarm of birds to flap through scrolling pipe gaps",
		Long: "Evolves a population (ga or neat) to fly birds through scrolling pipe gaps. " +
			"The window shows the whole swarm on one shared course, each bird flapping by its " +
			"own evolved network and crashing out until the fittest survive.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if p.agent != "ga" && p.agent != "neat" {
				return fmt.Errorf("flappy supports only the ga and neat agents")
			}
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			p.seed = resolveSeed(cmd, 0, log)
			flappyParams(&p)
			return runEnv(p, log)
		},
	}
	cmd.Flags().StringVar(&p.agent, "agent", p.agent, "agent: ga or neat")
	cmd.Flags().IntVar(&p.generations, "generations", p.generations, "generations to evolve")
	cmd.Flags().IntVar(&p.population, "population", p.population, "population size (birds in the swarm)")
	cmd.Flags().IntVar(&p.maxSteps, "max-steps", p.maxSteps, "maximum steps per episode")
	cmd.Flags().Int64Var(&seed, "seed", 0, "run seed (default: random, logged)")
	cmd.Flags().BoolVar(&p.headless, "headless", false, "train without a window")
	rootCmd.AddCommand(cmd)
}
