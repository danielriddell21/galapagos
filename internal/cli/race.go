package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/galapagos/internal/agents/ga"
	"github.com/danielriddell21/galapagos/internal/config"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/racing"
	"github.com/danielriddell21/galapagos/internal/gui"
	"github.com/danielriddell21/galapagos/internal/sim"
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
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			cfg := config.DefaultRacing()
			if cfgPath != "" {
				loaded, err := config.LoadRacing(cfgPath)
				if err != nil {
					return fmt.Errorf("load racing config: %w", err)
				}
				cfg = loaded
			}
			if cfg.Agent == "qlearning" {
				return fmt.Errorf("racing uses continuous control; choose agent ga or neat (qlearning runs on cartpole and maze)")
			}
			cfg.Seed = resolveSeed(cmd, cfg.Seed, log)
			if headless {
				return raceHeadless(cfg, out, log)
			}
			return raceGUI(cfg, out, log)
		},
	}
	cmd.Flags().StringVar(&cfgPath, "config", "", "path to a YAML config (defaults built in)")
	cmd.Flags().BoolVar(&headless, "headless", false, "train without a window, then save the best genome")
	cmd.Flags().StringVar(&out, "out", "best.json", "path to save the best genome")
	cmd.Flags().Int64Var(&seed, "seed", 0, "run seed (default: random, logged)")
	rootCmd.AddCommand(cmd)
}

func raceHeadless(c config.Racing, out string, log *slog.Logger) error {
	agent := buildAgent(c)
	rc := racingConfigFrom(c)
	tel := sim.NewTelemetry(c.Generations, log)

	log.Info("training", "agent", c.Agent, "generations", c.Generations, "population", c.Population, "seed", c.Seed)
	sim.TrainHeadless(envFactory(rc), agent, sim.RunConfig{
		Seed:        c.Seed,
		MaxSteps:    c.MaxSteps,
		Generations: c.Generations,
	}, tel)

	if pop, ok := agent.(*ga.Population); ok {
		if err := pop.SaveBest(out); err != nil {
			return fmt.Errorf("save best genome: %w", err)
		}
		fmt.Printf("saved best genome to %s\n", out)
	}
	return nil
}

func raceGUI(c config.Racing, out string, log *slog.Logger) error {
	if err := gui.Run(raceConfig(c, out), log); err != nil {
		return fmt.Errorf("run gui: %w", err)
	}
	return nil
}

// raceConfig wires the racing environment and its population into a run the
// display can drive. It is display-free, so the same wiring backs the window
// and the headless recordings in tools/demogen.
func raceConfig(c config.Racing, out string) gui.Config {
	env := racing.New(racingConfigFrom(c))
	agent := buildAgent(c)

	caps := runCaps{
		keymap: racingKeymap,
		bounds: boundsOf(env),
		leader: func(best int) (float64, float64, bool) {
			p := env.CarPosition(best)
			return p.X, p.Y, true
		},
		sensors: func(best int) (core.Vec2, []core.Vec2, bool) {
			o, ends := env.SensorEndpoints(best)
			return o, ends, true
		},
	}
	if gp, ok := agent.(*ga.Population); ok {
		caps.save = func() error {
			if err := gp.SaveBest(out); err != nil {
				return fmt.Errorf("save best genome: %w", err)
			}
			return nil
		}
		caps.load = func() error {
			sg, err := ga.LoadGenome(out)
			if err != nil {
				return fmt.Errorf("load genome: %w", err)
			}
			if err := gp.SetMemberGenome(0, sg.Genome); err != nil {
				return fmt.Errorf("set genome: %w", err)
			}
			return nil
		}
	}

	return populationGUI("Galapagos — race ("+c.Agent+")", env, agent, c.MaxSteps, c.Seed, caps)
}
