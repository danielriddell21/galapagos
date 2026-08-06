//go:build !ebiten

package gui

import (
	"errors"
	"log/slog"
)

func Available() bool { return false }

func Run(run Config, log *slog.Logger) error {
	// Recording needs no window: the software renderer draws the same frames
	// the window would, so demo media builds anywhere, with no display.
	if run.Rec.Recording() {
		return Render(run, log)
	}
	return errors.New("native window unavailable in this build: on macOS install the Homebrew cask, on other platforms run `galapagos serve` for the browser demo, or pass --headless")
}
