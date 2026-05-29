package main

import (
	"github.com/danielriddell21/galapagos/internal/agents/ga"
	"github.com/danielriddell21/galapagos/internal/agents/neat"
	"github.com/danielriddell21/galapagos/internal/config"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/racing"
	"github.com/danielriddell21/galapagos/internal/sim"
)

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
