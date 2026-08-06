package cli

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/cartpole"
)

// defaultCartpole is the cartpole command's flag defaults. The demo clip runs
// with them unchanged, so the recorded media shows what the command does.
func defaultCartpole() envParams {
	return envParams{agent: "ga", generations: 30, population: 60, episodes: 4000, maxSteps: 500, qbins: 8}
}

// cartpoleParams fills in the parts of p that describe the environment itself,
// once the flags (or a demo clip) have settled the tunables.
func cartpoleParams(p *envParams) {
	p.title = "Galapagos — cartpole (" + p.agent + ")"
	p.newSingle = func() core.Environment { return cartpole.New(cartpole.Config{MaxSteps: p.maxSteps}) }
}

func init() {
	p := defaultCartpole()
	var seed int64
	cmd := &cobra.Command{
		Use:   "cartpole",
		Short: "Balance a cart-pole with ga, neat, or q-learning",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			p.seed = resolveSeed(cmd, 0, log)
			cartpoleParams(&p)
			return runEnv(p, log)
		},
	}
	cmd.Flags().StringVar(&p.agent, "agent", p.agent, "agent: ga, neat, or qlearning")
	cmd.Flags().IntVar(&p.generations, "generations", p.generations, "generations (ga/neat)")
	cmd.Flags().IntVar(&p.population, "population", p.population, "population size (ga/neat)")
	cmd.Flags().IntVar(&p.episodes, "episodes", p.episodes, "episodes (qlearning)")
	cmd.Flags().IntVar(&p.maxSteps, "max-steps", p.maxSteps, "maximum steps per episode")
	cmd.Flags().Int64Var(&seed, "seed", 0, "run seed (default: random, logged)")
	cmd.Flags().BoolVar(&p.headless, "headless", false, "train without a window")
	rootCmd.AddCommand(cmd)
}
