package main

import (
	"fmt"
	"iter"

	"github.com/danielriddell21/galapagos/internal/agents/ga"
	"github.com/danielriddell21/galapagos/internal/config"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/racing"
	"github.com/danielriddell21/galapagos/internal/sim"
	"github.com/spf13/cobra"
)

func init() {
	var (
		genomePath string
		seed       int64
		maxSteps   int
	)
	cmd := &cobra.Command{
		Use:   "replay",
		Short: "Replay a saved genome on the track it was trained for",
		RunE: func(cmd *cobra.Command, args []string) error {
			sg, err := ga.LoadGenome(genomePath)
			if err != nil {
				return err
			}
			rc := racing.DefaultConfig()
			rc.Sensors.Count = sg.Inputs - 1
			reward := replayDriver(rc, ga.NewDriver(sg), seed, maxSteps)
			fmt.Printf("replayed genome from generation %d: reward %.3f\n", sg.Generation, reward)
			return nil
		},
	}
	def := config.DefaultRacing()
	cmd.Flags().StringVar(&genomePath, "genome", "best.json", "path to a saved genome")
	cmd.Flags().Int64Var(&seed, "seed", def.Seed, "track seed to replay on")
	cmd.Flags().IntVar(&maxSteps, "max-steps", def.MaxSteps, "maximum steps to simulate")
	rootCmd.AddCommand(cmd)
}

// replayDriver rolls a single driver out on a fresh track and returns its total
// reward. It reuses the training loop with a population of one, so replay is
// deterministic and matches the trained result.
func replayDriver(rc racing.Config, d core.Individual, seed int64, maxSteps int) float64 {
	fitness := sim.RunGeneration(racing.New(rc), singleDriverPop{d}, maxSteps, seed, nil)
	return fitness[0]
}

// singleDriverPop adapts one driver to the PopulationAgent interface so it can
// be replayed through the standard simulation loop.
type singleDriverPop [1]core.Individual

func (p singleDriverPop) Act(s core.State) core.Action { return p[0].Act(s) }
func (p singleDriverPop) Observe(s core.State, a core.Action, r core.Reward, n core.State, d bool) {
}
func (p singleDriverPop) EndEpisode(total core.Reward) {}
func (p singleDriverPop) Len() int                     { return 1 }
func (p singleDriverPop) Generation() int              { return 0 }
func (p singleDriverPop) Evolve()                      {}
func (p singleDriverPop) All() iter.Seq[core.Individual] {
	return func(yield func(core.Individual) bool) { yield(p[0]) }
}

var _ core.PopulationAgent = singleDriverPop{}
