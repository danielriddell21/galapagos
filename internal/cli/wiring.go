package cli

import (
	"fmt"
	"log/slog"
	"math/rand/v2"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/galapagos/internal/agents/ga"
	"github.com/danielriddell21/galapagos/internal/agents/neat"
	"github.com/danielriddell21/galapagos/internal/config"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/racing"
	"github.com/danielriddell21/galapagos/internal/sim"
)

// keymaps shown in the GUI's top-right overlay, per kind of run.
var (
	racingKeymap = []string{"space pause", "f follow", "+/- speed", "r new track", "d rays", "s save", "l load"}
	swarmKeymap  = []string{"space pause", "+/- speed", "r regenerate"}
	onlineKeymap = []string{"space pause", "+/- speed", "r regenerate"}
)

// resolveSeed chooses the run seed: an explicit --seed wins; otherwise a
// non-zero config seed; otherwise a fresh random seed. The chosen seed is always
// logged so a random run can be reproduced with --seed.
func resolveSeed(cmd *cobra.Command, configSeed int64, log *slog.Logger) int64 {
	var seed int64
	switch {
	case cmd.Flags().Changed("seed"):
		seed, _ = cmd.Flags().GetInt64("seed")
	case configSeed != 0:
		seed = configSeed
	default:
		seed = int64(rand.Uint64() >> 1) // random, non-negative
	}
	log.Info("seed", "value", seed)
	return seed
}

// discreteCount returns the number of discrete actions described by a spec, or 0
// when the action is continuous.
func discreteCount(s core.Spec) int {
	if !s.Discrete || len(s.High) == 0 {
		return 0
	}
	return int(s.High[0]-s.Low[0]) + 1
}

// newPopulationAgent builds GA or NEAT sized to an environment's specs.
func newPopulationAgent(name string, obs, act core.Spec, population int, seed int64) (core.PopulationAgent, error) {
	switch name {
	case "ga":
		return ga.New(ga.Config{
			Population: population, EliteFraction: 0.1, MutationRate: 0.05, MutationStd: 0.2,
			HiddenSize: 8, Inputs: obs.Dim, Outputs: act.Dim, Seed: seed,
		}), nil
	case "neat":
		nc := neat.DefaultConfig()
		nc.Population, nc.Inputs, nc.Outputs, nc.Seed = population, obs.Dim, act.Dim, seed
		return neat.New(nc), nil
	default:
		return nil, fmt.Errorf("unknown population agent %q (want ga or neat)", name)
	}
}

// gaConfigFrom builds the genetic-algorithm config from the run config. The
// network input size is the ray count plus one for normalized speed; the output
// size is two (steering and throttle).
func gaConfigFrom(c config.Racing) ga.Config {
	return ga.Config{
		Population:    c.Population,
		EliteFraction: c.EliteFraction,
		MutationRate:  c.MutationRate,
		MutationStd:   c.MutationStd,
		HiddenSize:    c.HiddenSize,
		Inputs:        c.Rays + 1,
		Outputs:       2,
		Seed:          c.Seed,
	}
}

// racingConfigFrom builds the environment config from the run config.
func racingConfigFrom(c config.Racing) racing.Config {
	rc := racing.DefaultConfig()
	rc.Sensors.Count = c.Rays
	return rc
}

// neatConfigFrom builds a NEAT config from the run config.
func neatConfigFrom(c config.Racing) neat.Config {
	nc := neat.DefaultConfig()
	nc.Population = c.Population
	nc.Inputs = c.Rays + 1
	nc.Outputs = 2
	nc.Seed = c.Seed
	return nc
}

// buildAgent constructs the configured population-based agent.
func buildAgent(c config.Racing) core.PopulationAgent {
	if c.Agent == "neat" {
		return neat.New(neatConfigFrom(c))
	}
	return ga.New(gaConfigFrom(c))
}

// envFactory returns a factory that builds independent racing environments,
// used for parallel evaluation.
func envFactory(rc racing.Config) sim.EnvFactory {
	return func() core.MultiEnvironment { return racing.New(rc) }
}
