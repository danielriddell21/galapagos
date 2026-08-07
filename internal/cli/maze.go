package cli

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/maze"
)

// defaultMazeSize is the maze command's default width and height in cells. It
// also sets the q-learning bin count, so the table is keyed by whole cells.
const defaultMazeSize = 6

// defaultMaze is the maze command's flag defaults. The demo clip runs with
// them unchanged, so the recorded media shows what the command does.
func defaultMaze() envParams {
	return envParams{agent: "qlearning", generations: 30, population: 60, episodes: 3000, maxSteps: 300}
}

// mazeParams fills in the parts of p that describe the environment itself,
// once the flags (or a demo clip) have settled the tunables.
func mazeParams(p *envParams, size int) {
	p.title = "Galapagos — maze (" + p.agent + ")"
	p.qbins = size
	p.newSingle = func() core.Environment {
		return maze.New(maze.Config{Width: size, Height: size, MaxSteps: p.maxSteps})
	}
}

func init() {
	p := defaultMaze()
	var (
		size int
		seed int64
	)
	cmd := &cobra.Command{
		Use:   "maze",
		Short: "Solve a grid maze with q-learning, ga, or neat",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			p.seed = resolveSeed(cmd, 0, log)
			mazeParams(&p, size)
			return runEnv(p, log)
		},
	}
	cmd.Flags().StringVar(&p.agent, "agent", p.agent, "agent: qlearning, ga, or neat")
	cmd.Flags().IntVar(&size, "size", defaultMazeSize, "maze width and height in cells")
	cmd.Flags().IntVar(&p.episodes, "episodes", p.episodes, "episodes (qlearning)")
	cmd.Flags().IntVar(&p.generations, "generations", p.generations, "generations (ga/neat)")
	cmd.Flags().IntVar(&p.population, "population", p.population, "population size (ga/neat)")
	cmd.Flags().IntVar(&p.maxSteps, "max-steps", p.maxSteps, "maximum steps per episode")
	cmd.Flags().Int64Var(&seed, "seed", 0, "run seed (default: random, logged)")
	cmd.Flags().BoolVar(&p.headless, "headless", false, "train without a window")
	rootCmd.AddCommand(cmd)
}
