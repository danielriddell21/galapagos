package cli

import (
	"fmt"
	"log/slog"

	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/galapagos/internal/config"
	"github.com/danielriddell21/galapagos/internal/gui"
)

// DemoSpec is one documentation clip: which environment to wire up, which
// agent drives it, and the seed that makes the run reproducible. Everything
// else — population sizes, episode lengths, training budgets — comes from the
// same defaults the command registers, so the media shows what a reader gets
// by running the command.
type DemoSpec struct {
	// Env names the environment: race, cartpole, maze, cube, flappy or chess.
	Env string
	// Agent overrides the environment's default agent when set.
	Agent string
	// Opponent overrides the chess opponent when set; other environments
	// ignore it.
	Opponent string
	// Seed makes the run reproducible.
	Seed int64
	// Rec says where to write the recording and how long it runs.
	Rec record.Options
}

// DemoEnvs lists the environments [Demo] can record, in the order the
// documentation presents them.
func DemoEnvs() []string {
	return []string{"race", "cartpole", "maze", "cube", "flappy", "chess"}
}

// Demo records one documentation clip without opening a window. It wires the
// environment exactly as the matching command does, then hands the run to the
// software renderer, so the media matches what a player sees and needs no
// display.
//
// Environments that must learn before there is anything to watch — cube trains
// a policy, chess evolves a population — do that first, so a clip can take a
// while.
func Demo(spec DemoSpec, log *slog.Logger) error {
	cfg, err := demoConfig(spec, log)
	if err != nil {
		return err
	}
	cfg.Rec = spec.Rec
	if err := gui.Render(cfg, log); err != nil {
		return fmt.Errorf("render %s: %w", spec.Env, err)
	}
	return nil
}

func demoConfig(spec DemoSpec, log *slog.Logger) (gui.Config, error) {
	switch spec.Env {
	case "race":
		c := config.DefaultRacing()
		if spec.Agent != "" {
			c.Agent = spec.Agent
		}
		c.Seed = spec.Seed
		return raceConfig(c, "best.json"), nil

	case "cartpole":
		p := defaultCartpole()
		return demoEnvConfig(&p, spec, cartpoleParams)

	case "flappy":
		p := defaultFlappy()
		return demoEnvConfig(&p, spec, flappyParams)

	case "maze":
		p := defaultMaze()
		return demoEnvConfig(&p, spec, func(p *envParams) { mazeParams(p, defaultMazeSize) })

	case "cube":
		c := defaultCube()
		c.seed = spec.Seed
		policy, beam, err := c.solve(log)
		if err != nil {
			return gui.Config{}, err
		}
		return cubeConfig(policy, beam, c.cubes, c.guiDepth, c.seed), nil

	case "chess":
		c := defaultChess()
		if spec.Agent != "" {
			c.agent = spec.Agent
		}
		if spec.Opponent != "" {
			c.opponent = spec.Opponent
		}
		c.seed = spec.Seed
		pop, _, err := evolveChess(c, log)
		if err != nil {
			return gui.Config{}, err
		}
		return chessConfig(pop, c.maxPlies, c.seed, c.agent)

	default:
		return gui.Config{}, fmt.Errorf("unknown demo environment %q (want one of %v)", spec.Env, DemoEnvs())
	}
}

// demoEnvConfig applies a clip's overrides to an environment's defaults and
// wires the run.
func demoEnvConfig(p *envParams, spec DemoSpec, describe func(*envParams)) (gui.Config, error) {
	if spec.Agent != "" {
		p.agent = spec.Agent
	}
	p.seed = spec.Seed
	describe(p)
	return envConfig(*p)
}
