//go:build ebiten

package gui

import (
	"fmt"
	"image"
	"log/slog"
	"math/rand/v2"
	"time"

	eb "github.com/hajimehoshi/ebiten/v2"
	ebinput "github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/danielriddell21/crucible/record"
	"github.com/danielriddell21/crucible/window"

	ebrender "github.com/danielriddell21/galapagos/internal/render/ebiten"
)

func Available() bool { return true }

func randomSeed() int64 { return int64(rand.Uint64() >> 1) }

type game struct {
	run   Config
	log   *slog.Logger
	ren   *ebrender.Renderer
	start time.Time

	paused   bool
	follow   bool
	showRays bool
	speed    int

	rec   *record.Recorder
	pix   []byte
	saved bool
}

func Run(run Config, log *slog.Logger) error {
	g := &game{
		run:   run,
		log:   log,
		ren:   ebrender.New(screenW, screenH),
		start: time.Now(),
		speed: 1,
	}
	if run.Rec.Recording() {
		g.rec = record.New(run.Rec)
		g.speed = 2 // a steady pace for a lively recording
	}
	window.Configure(window.Options{Title: run.Title, Width: screenW, Height: screenH, MinWidth: screenW / 2, MinHeight: screenH / 2})
	if err := eb.RunGame(g); err != nil {
		return fmt.Errorf("run window: %w", err)
	}
	return nil
}

func (g *game) Update() error {
	if g.rec != nil && g.rec.Done() {
		if !g.saved {
			if err := g.rec.Save(g.run.Rec.Path); err != nil {
				g.log.Error("record failed", "err", err)
			} else {
				g.log.Info("recorded", "path", g.run.Rec.Path, "frames", g.rec.Len())
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
		if g.run.Step() {
			if g.run.Complete != nil {
				g.run.Complete()
			}
			g.run.Next()
		}
	}
	return nil
}

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
		g.run.Regenerate(seed)
	case ebinput.IsKeyJustPressed(eb.KeyS):
		if g.run.Save != nil {
			if err := g.run.Save(); err != nil {
				g.log.Error("save failed", "err", err)
			} else {
				g.log.Info("saved")
			}
		}
	case ebinput.IsKeyJustPressed(eb.KeyL):
		if g.run.Load != nil {
			if err := g.run.Load(); err != nil {
				g.log.Error("load failed", "err", err)
			} else {
				g.log.Info("loaded")
			}
		}
	}
}

func (g *game) Draw(screen *eb.Image) {
	screen.Fill(Background)
	g.ren.Begin(screen)

	DrawFrame(g.ren, g.run, FrameState{
		Speed:    g.speed,
		Paused:   g.paused,
		Elapsed:  time.Since(g.start),
		ShowRays: g.showRays,
		Follow:   g.follow,
	})

	if g.rec != nil && !g.rec.Done() {
		if g.pix == nil {
			g.pix = make([]byte, 4*screenW*screenH)
		}
		screen.ReadPixels(g.pix)
		g.rec.Add(&image.RGBA{Pix: g.pix, Stride: 4 * screenW, Rect: image.Rect(0, 0, screenW, screenH)})
	}
}

func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenW, screenH
}
