package cli

import (
	"fmt"

	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/gui"
	"github.com/danielriddell21/galapagos/internal/sim"
)

type runCaps struct {
	keymap  []string
	bounds  func() (minX, minY, maxX, maxY float64, ok bool)
	leader  func(best int) (x, y float64, ok bool)
	sensors func(best int) (origin core.Vec2, ends []core.Vec2, ok bool)
	save    func() error
	load    func() error
}

func populationGUI(title string, env core.MultiEnvironment, pop core.PopulationAgent, maxSteps int, seed int64, caps runCaps) gui.Config {
	live := sim.NewLive(env, pop, maxSteps, seed)
	tel := sim.NewTelemetry(4096, nil)
	best := live.BestIndex

	cfg := gui.Config{Title: title, Keymap: caps.keymap, Population: true}
	cfg.Step = live.Step
	cfg.Complete = func() { tel.Publish(sim.StatsFrom(pop.Generation(), live.Fitness())) }
	cfg.Next = live.NextGeneration
	cfg.Render = func(r core.Renderer) {
		if sb, ok := env.(interface{ SetBest(int) }); ok {
			sb.SetBest(best())
		}
		env.Render(r)
	}
	cfg.HUD = func() []string {
		f := live.Fitness()
		b := best()
		lines := []string{
			fmt.Sprintf("generation %d", pop.Generation()),
			fmt.Sprintf("alive %d/%d", env.Alive(), pop.Len()),
			fmt.Sprintf("best %.1f", f[b]),
		}
		if s, ok := env.(interface{ BodyStatus(int) (string, bool) }); ok {
			if line, ok2 := s.BodyStatus(b); ok2 {
				lines = append(lines, line)
			}
		}
		return lines
	}
	cfg.Series = tel.BestSeries
	cfg.Bounds = caps.bounds
	cfg.Leader = func() (float64, float64, bool) { return caps.leader(best()) }
	cfg.Sensors = func() (core.Vec2, []core.Vec2, bool) { return caps.sensors(best()) }
	cfg.Save = caps.save
	cfg.Load = caps.load
	cfg.Regenerate = live.Regenerate
	return cfg
}

func onlineGUI(title string, env core.Environment, agent core.Agent, maxSteps int, seed int64, caps runCaps) gui.Config {
	ep := sim.NewLiveEpisode(env, agent, maxSteps, seed)
	var returns []float64
	solver, goalBased := env.(interface{ Solved() bool })
	solved := 0

	cfg := gui.Config{Title: title, Keymap: caps.keymap}
	cfg.Step = ep.Step
	cfg.Complete = func() {
		returns = append(returns, float64(ep.Return()))
		if goalBased && solver.Solved() {
			solved++
		}
	}
	cfg.Next = ep.NextEpisode
	cfg.Render = func(r core.Renderer) { env.Render(r) }
	cfg.HUD = func() []string {
		lines := []string{
			fmt.Sprintf("episode %d", ep.Episode()),
			fmt.Sprintf("step %d", ep.StepCount()),
			fmt.Sprintf("return %.2f", float64(ep.Return())),
		}
		if goalBased {
			lines = append(lines, fmt.Sprintf("solved %d", solved))
		}
		return lines
	}
	cfg.Series = func() []float64 { return returns }
	cfg.Bounds = caps.bounds
	cfg.Leader = func() (float64, float64, bool) { return 0, 0, false }
	cfg.Sensors = func() (core.Vec2, []core.Vec2, bool) { return core.Vec2{}, nil, false }
	cfg.Regenerate = ep.Regenerate
	return cfg
}

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

func noLeader(int) (float64, float64, bool)        { return 0, 0, false }
func noSensors(int) (core.Vec2, []core.Vec2, bool) { return core.Vec2{}, nil, false }
