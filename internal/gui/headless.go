package gui

import (
	"fmt"
	"image"
	"log/slog"
	"time"

	"github.com/danielriddell21/crucible/demo"
	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/galapagos/internal/render/soft"
)

// recordSpeed advances the simulation this many steps per captured frame, the
// same steady pace the windowed recorder uses.
const recordSpeed = 2

// simElapsed converts a frame index into the run time the HUD should show. A
// headless run has no wall clock worth reporting, so it reports the time the
// finished recording will have played by that frame.
func simElapsed(step int) time.Duration {
	fps := max(defaultRecordFPS, 1)
	return time.Duration(step) * time.Second / time.Duration(fps)
}

// defaultRecordFPS matches the playback rate galapagos records at.
const defaultRecordFPS = 25

// Render records a run to cfg.Rec.Path without opening a window, drawing every
// frame through the software renderer. It composes frames with the same
// [DrawFrame] the window uses, so the media matches what a player sees, and it
// needs no display — a docs build can run it anywhere.
//
// The file extension picks the format: .gif or .mp4.
func Render(cfg Config, log *slog.Logger) error {
	r := soft.New(screenW, screenH)
	rec := record.New(cfg.Rec)

	clip := demo.Clip{
		Frames: cfg.Rec.Frames,
		Step: func(int) error {
			for range recordSpeed {
				if cfg.Step() {
					if cfg.Complete != nil {
						cfg.Complete()
					}
					cfg.Next()
				}
			}
			return nil
		},
		Frame: func(step int) image.Image {
			r.Clear(Background)
			DrawFrame(r, cfg, FrameState{
				Speed: recordSpeed,
				// The elapsed clock is wall-time in the window; headless runs
				// report simulated time so the media is reproducible.
				Elapsed: simElapsed(step),
			})
			return r.Image()
		},
	}
	if _, err := clip.Record(rec); err != nil {
		return fmt.Errorf("capture run: %w", err)
	}
	if err := rec.Save(cfg.Rec.Path); err != nil {
		return fmt.Errorf("save recording: %w", err)
	}
	if log != nil {
		log.Info("recorded", "path", cfg.Rec.Path, "frames", rec.Len())
	}
	return nil
}
