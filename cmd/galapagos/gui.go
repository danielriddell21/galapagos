package main

import (
	"fmt"

	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/sim"
)

// guiRun is a backend-agnostic description of a windowed run: a bundle of
// closures the Ebiten game drives. Building it in plain code keeps the GUI from
// importing any concrete environment or agent, and lets the same window serve
// every env/agent combination.
type guiRun struct {
	title      string
	keymap     []string
	population bool // population run (generations) vs online run (episodes)

	step     func() bool         // advance one tick; true when the gen/episode ends
	complete func()              // called once when a gen/episode ends, before next
	next     func()              // evolve / start the next episode
	render   func(core.Renderer) // draw the world
	hud      func() []string     // HUD lines, top-left
	series   func() []float64    // fitness/return history for the sparkline

	bounds  func() (minX, minY, maxX, maxY float64, ok bool)
	leader  func() (x, y float64, ok bool)                       // follow target
	sensors func() (origin core.Vec2, ends []core.Vec2, ok bool) // ray overlay

	save       func() error
	load       func() error
	regenerate func(seed int64)
}

// runCaps holds the environment-specific capabilities a command supplies when
// assembling a run.
type runCaps struct {
	keymap  []string
	bounds  func() (minX, minY, maxX, maxY float64, ok bool)
	leader  func(best int) (x, y float64, ok bool)
	sensors func(best int) (origin core.Vec2, ends []core.Vec2, ok bool)
	save    func() error
	load    func() error
}

// populationGUI assembles a run that evolves a population on a multi-environment,
// driven frame by frame by sim.Live.
func populationGUI(title string, env core.MultiEnvironment, pop core.PopulationAgent, maxSteps int, seed int64, caps runCaps) guiRun {
	live := sim.NewLive(env, pop, maxSteps, seed)
	tel := sim.NewTelemetry(4096, nil)
	best := live.BestIndex

	return guiRun{
		title:      title,
		keymap:     caps.keymap,
		population: true,
		step:       live.Step,
		complete:   func() { tel.Publish(sim.StatsFrom(pop.Generation(), live.Fitness())) },
		next:       live.NextGeneration,
		render: func(r core.Renderer) {
			if sb, ok := env.(interface{ SetBest(int) }); ok {
				sb.SetBest(best())
			}
			env.Render(r)
		},
		hud: func() []string {
			f := live.Fitness()
			return []string{
				fmt.Sprintf("generation %d", pop.Generation()),
				fmt.Sprintf("alive %d/%d", env.Alive(), pop.Len()),
				fmt.Sprintf("best %.1f", f[best()]),
			}
		},
		series:     tel.BestSeries,
		bounds:     caps.bounds,
		leader:     func() (float64, float64, bool) { return caps.leader(best()) },
		sensors:    func() (core.Vec2, []core.Vec2, bool) { return caps.sensors(best()) },
		save:       caps.save,
		load:       caps.load,
		regenerate: live.Regenerate,
	}
}

// onlineGUI assembles a run that trains a single online agent on an environment,
// one episode at a time, driven by sim.LiveEpisode.
func onlineGUI(title string, env core.Environment, agent core.Agent, maxSteps int, seed int64, caps runCaps) guiRun {
	ep := sim.NewLiveEpisode(env, agent, maxSteps, seed)
	var returns []float64

	return guiRun{
		title:    title,
		keymap:   caps.keymap,
		step:     ep.Step,
		complete: func() { returns = append(returns, float64(ep.Return())) },
		next:     ep.NextEpisode,
		render:   func(r core.Renderer) { env.Render(r) },
		hud: func() []string {
			return []string{
				fmt.Sprintf("episode %d", ep.Episode()),
				fmt.Sprintf("step %d", ep.StepCount()),
				fmt.Sprintf("return %.2f", float64(ep.Return())),
			}
		},
		series:     func() []float64 { return returns },
		bounds:     caps.bounds,
		leader:     func() (float64, float64, bool) { return 0, 0, false },
		sensors:    func() (core.Vec2, []core.Vec2, bool) { return core.Vec2{}, nil, false },
		regenerate: ep.Regenerate,
	}
}

// boundsOf returns a bounds closure for any environment that exposes a Bounds
// method (single-agent envs return four values; the multi-adapter returns five).
func boundsOf(env any) func() (float64, float64, float64, float64, bool) {
	if b, ok := env.(interface {
		Bounds() (float64, float64, float64, float64)
	}); ok {
		return func() (float64, float64, float64, float64, bool) {
			x0, y0, x1, y1 := b.Bounds()
			return x0, y0, x1, y1, true
		}
	}
	if b, ok := env.(interface {
		Bounds() (float64, float64, float64, float64, bool)
	}); ok {
		return b.Bounds
	}
	return func() (float64, float64, float64, float64, bool) { return 0, 0, 0, 0, false }
}

// noLeader and noSensors are default capabilities for environments without a
// followable leader or sensor overlay.
func noLeader(int) (float64, float64, bool)        { return 0, 0, false }
func noSensors(int) (core.Vec2, []core.Vec2, bool) { return core.Vec2{}, nil, false }
