//go:build ebiten

package cli

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"math/rand/v2"
	"time"

	eb "github.com/hajimehoshi/ebiten/v2"
	ebinput "github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/danielriddell21/galapagos/internal/core"
	ebrender "github.com/danielriddell21/galapagos/internal/render/ebiten"
)

// randomSeed returns a fresh non-negative seed for the GUI's regenerate control.
func randomSeed() int64 { return int64(rand.Uint64() >> 1) }

const (
	screenW = 1024
	screenH = 768
)

// game is the generic Ebiten game: it drives any guiRun, so the same window
// serves every environment and agent.
type game struct {
	run   guiRun
	log   *slog.Logger
	ren   *ebrender.Renderer
	start time.Time

	paused   bool
	follow   bool
	showRays bool
	speed    int

	rec   *recorder // non-nil when recording a GIF
	pix   []byte    // reusable RGBA readback buffer
	saved bool
}

// launchGUI opens a window driving the given run. When --record is set it records
// a GIF and exits instead of running interactively.
func launchGUI(run guiRun, log *slog.Logger) error {
	g := &game{
		run:   run,
		log:   log,
		ren:   ebrender.New(screenW, screenH),
		start: time.Now(),
		speed: 1,
	}
	if recordPath != "" {
		g.rec = newRecorder(recordFrames, recordFPS, recordScale)
		g.speed = 2 // a steady pace for a lively recording
	}
	eb.SetWindowSize(screenW, screenH)
	eb.SetWindowTitle(run.title)
	if err := eb.RunGame(g); err != nil {
		return fmt.Errorf("run window: %w", err)
	}
	return nil
}

// Update handles input and advances the simulation. When recording finishes it
// writes the GIF and terminates the game loop.
func (g *game) Update() error {
	if g.rec != nil && g.rec.done {
		if !g.saved {
			if err := g.rec.save(recordPath); err != nil {
				g.log.Error("record failed", "err", err)
			} else {
				g.log.Info("recorded", "path", recordPath, "frames", len(g.rec.frames))
			}
			g.saved = true
		}
		return eb.Termination
	}
	g.handleInput()
	if g.paused {
		return nil
	}
	for range g.speed {
		if g.run.step() {
			if g.run.complete != nil {
				g.run.complete()
			}
			g.run.next()
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
	case ebinput.IsKeyJustPressed(eb.KeyEqual):
		g.speed = min(g.speed*2, 256)
	case ebinput.IsKeyJustPressed(eb.KeyMinus):
		g.speed = max(g.speed/2, 1)
	case ebinput.IsKeyJustPressed(eb.KeyD):
		g.showRays = !g.showRays
	case ebinput.IsKeyJustPressed(eb.KeyR):
		seed := randomSeed()
		g.log.Info("regenerate", "seed", seed)
		g.run.regenerate(seed)
	case ebinput.IsKeyJustPressed(eb.KeyS):
		if g.run.save != nil {
			if err := g.run.save(); err != nil {
				g.log.Error("save failed", "err", err)
			} else {
				g.log.Info("saved")
			}
		}
	case ebinput.IsKeyJustPressed(eb.KeyL):
		if g.run.load != nil {
			if err := g.run.load(); err != nil {
				g.log.Error("load failed", "err", err)
			} else {
				g.log.Info("loaded")
			}
		}
	}
}

// Draw renders the world, HUD, sparkline, and keymap overlay.
func (g *game) Draw(screen *eb.Image) {
	screen.Fill(color.RGBA{18, 20, 26, 255})
	g.ren.Begin(screen)

	g.updateCamera()
	g.run.render(g.ren)
	if g.showRays {
		if origin, ends, ok := g.run.sensors(); ok {
			for _, e := range ends {
				g.ren.Line(origin.X, origin.Y, e.X, e.Y, color.RGBA{0, 200, 120, 160})
			}
		}
	}
	g.drawHUD()
	g.drawKeymap()
	g.drawSparkline(g.run.series())

	if g.rec != nil && !g.rec.done {
		if g.pix == nil {
			g.pix = make([]byte, 4*screenW*screenH)
		}
		screen.ReadPixels(g.pix)
		g.rec.add(&image.RGBA{Pix: g.pix, Stride: 4 * screenW, Rect: image.Rect(0, 0, screenW, screenH)})
	}
}

// updateCamera follows the leader when enabled and available, otherwise frames
// the whole world.
func (g *game) updateCamera() {
	cam := g.ren.Camera()
	if g.follow {
		if x, y, ok := g.run.leader(); ok {
			cam.Zoom = 1.2
			cam.Follow(x, y)
			return
		}
	}
	if minX, minY, maxX, maxY, ok := g.run.bounds(); ok {
		cam.FitBounds(minX, minY, maxX, maxY, 0.15)
	}
}

// drawHUD draws the run's HUD lines plus speed and elapsed time, top-left.
func (g *game) drawHUD() {
	y := 10
	for _, line := range g.run.hud() {
		g.ren.Text(10, float64(y), line)
		y += 16
	}
	g.ren.Text(10, float64(y), fmt.Sprintf("speed x%d%s", g.speed, pausedLabel(g.paused)))
	g.ren.Text(10, float64(y+16), fmt.Sprintf("elapsed %s", time.Since(g.start).Round(time.Second)))
}

// drawKeymap lists the controls in the top-right corner.
func (g *game) drawKeymap() {
	y := 10
	for _, line := range g.run.keymap {
		g.ren.Text(screenW-150, float64(y), line)
		y += 16
	}
}

// drawSparkline plots the fitness/return history in screen space, bottom-left.
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
	// Centering the camera on the viewport with unit zoom makes world == screen
	// coordinates, so the HUD draws in screen space.
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
