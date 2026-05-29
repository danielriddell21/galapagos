//go:build ebiten

package main

import (
	"fmt"
	"image/color"
	"log/slog"
	"os"
	"time"

	eb "github.com/hajimehoshi/ebiten/v2"
	ebinput "github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/danielriddell21/galapagos/internal/agents/ga"
	"github.com/danielriddell21/galapagos/internal/config"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/racing"
	ebrender "github.com/danielriddell21/galapagos/internal/render/ebiten"
	"github.com/danielriddell21/galapagos/internal/sim"
)

const (
	screenW = 1024
	screenH = 768
)

// game is the Ebiten game driving the live racing demo.
type game struct {
	cfg  config.Racing
	out  string
	log  *slog.Logger
	ren  *ebrender.Renderer
	env  *racing.Env
	pop  *ga.Population
	live *sim.Live
	tel  *sim.Telemetry

	paused   bool
	follow   bool
	showRays bool
	speed    int   // steps simulated per frame
	seed     int64 // current track seed
	start    time.Time
}

// launchGUI runs the windowed demo: a live, evolving population on the track.
func launchGUI(c config.Racing, out string) error {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	env := racing.New(racingConfigFrom(c))
	pop := ga.New(gaConfigFrom(c))
	g := &game{
		cfg:   c,
		out:   out,
		log:   log,
		ren:   ebrender.New(screenW, screenH),
		env:   env,
		pop:   pop,
		live:  sim.NewLive(env, pop, c.MaxSteps, c.Seed),
		tel:   sim.NewTelemetry(c.Generations, log),
		speed: 1,
		seed:  c.Seed,
		start: time.Now(),
	}
	eb.SetWindowSize(screenW, screenH)
	eb.SetWindowTitle("Galapagos — race")
	if err := eb.RunGame(g); err != nil {
		return fmt.Errorf("run window: %w", err)
	}
	return nil
}

// Update advances input handling and the simulation.
func (g *game) Update() error {
	g.handleInput()
	if g.paused {
		return nil
	}
	for range g.speed {
		if g.live.Step() {
			g.tel.Publish(sim.StatsFrom(g.pop.Generation(), g.live.Fitness()))
			g.live.NextGeneration()
		}
	}
	return nil
}

// handleInput maps the documented controls to actions.
func (g *game) handleInput() {
	switch {
	case ebinput.IsKeyJustPressed(eb.KeySpace):
		g.paused = !g.paused
	case ebinput.IsKeyJustPressed(eb.KeyF):
		g.follow = !g.follow
	case ebinput.IsKeyJustPressed(eb.KeyEqual): // '+'
		g.speed = min(g.speed*2, 256)
	case ebinput.IsKeyJustPressed(eb.KeyMinus):
		g.speed = max(g.speed/2, 1)
	case ebinput.IsKeyJustPressed(eb.KeyR):
		g.seed++
		g.live.Regenerate(g.seed)
	case ebinput.IsKeyJustPressed(eb.KeyD):
		g.showRays = !g.showRays
	case ebinput.IsKeyJustPressed(eb.KeyS):
		if err := g.pop.SaveBest(g.out); err != nil {
			g.log.Error("save failed", "err", err)
		} else {
			g.log.Info("saved best genome", "path", g.out)
		}
	case ebinput.IsKeyJustPressed(eb.KeyL):
		if sg, err := ga.LoadGenome(g.out); err != nil {
			g.log.Error("load failed", "err", err)
		} else if err := g.pop.SetMemberGenome(0, sg.Genome); err != nil {
			g.log.Error("inject failed", "err", err)
		} else {
			g.log.Info("loaded best genome into member 0", "path", g.out)
		}
	}
}

// Draw renders the world and HUD.
func (g *game) Draw(screen *eb.Image) {
	screen.Fill(color.RGBA{18, 20, 26, 255})
	g.ren.Begin(screen)

	best := g.live.BestIndex()
	g.updateCamera(best)
	g.env.SetBest(best)
	g.env.Render(g.ren)
	if g.showRays {
		g.drawRays(best)
	}
	g.drawHUD(best)
}

// updateCamera follows the leader or frames the whole track.
func (g *game) updateCamera(best int) {
	cam := g.ren.Camera()
	if g.follow {
		p := g.env.CarPosition(best)
		cam.Zoom = 1.2
		cam.Follow(p.X, p.Y)
		return
	}
	if t := g.env.Track(); t != nil {
		minX, minY, maxX, maxY := t.Bounds()
		cam.FitBounds(minX, minY, maxX, maxY, 0.15)
	}
}

// drawRays overlays the best car's sensor rays.
func (g *game) drawRays(best int) {
	origin, ends := g.env.SensorEndpoints(best)
	for _, e := range ends {
		g.ren.Line(origin.X, origin.Y, e.X, e.Y, color.RGBA{0, 200, 120, 160})
	}
}

// drawHUD draws the generation summary and a fitness-over-time sparkline.
func (g *game) drawHUD(best int) {
	fit := g.live.Fitness()
	g.ren.Text(10, 10, fmt.Sprintf("generation %d", g.pop.Generation()))
	g.ren.Text(10, 26, fmt.Sprintf("alive %d / %d", g.env.Alive(), g.pop.Len()))
	g.ren.Text(10, 42, fmt.Sprintf("best %.1f", fit[best]))
	g.ren.Text(10, 58, fmt.Sprintf("speed x%d%s", g.speed, pausedLabel(g.paused)))
	g.ren.Text(10, 74, fmt.Sprintf("elapsed %s", time.Since(g.start).Round(time.Second)))
	g.drawSparkline(g.tel.BestSeries())
}

// drawSparkline plots the best-fitness series in screen space.
func (g *game) drawSparkline(series []float64) {
	if len(series) < 2 {
		return
	}
	const x0, y0, w, h = 10.0, 700.0, 240.0, 50.0
	lo, hi := series[0], series[0]
	for _, v := range series {
		lo, hi = min(lo, v), max(hi, v)
	}
	span := max(hi-lo, 1e-9)

	cam := g.ren.Camera()
	prev := *cam
	// A camera centered on the viewport with unit zoom makes world coordinates
	// equal screen pixels, so the HUD draws in screen space.
	*cam = core.Camera{X: screenW / 2, Y: screenH / 2, Zoom: 1, ViewW: screenW, ViewH: screenH}
	for i := 1; i < len(series); i++ {
		ax := x0 + w*float64(i-1)/float64(len(series)-1)
		bx := x0 + w*float64(i)/float64(len(series)-1)
		ay := y0 + h - h*(series[i-1]-lo)/span
		by := y0 + h - h*(series[i]-lo)/span
		g.ren.Line(ax, ay, bx, by, color.RGBA{255, 215, 0, 255})
	}
	*cam = prev
}

func pausedLabel(p bool) string {
	if p {
		return " (paused)"
	}
	return ""
}

// Layout implements ebiten.Game.
func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenW, screenH
}
