//go:build !ebiten

package gui

import (
	"errors"
	"log/slog"
)

func Available() bool { return false }

func Run(_ Config, _ *slog.Logger) error {
	return errors.New("native window unavailable in this build: on macOS install the Homebrew cask, pass --headless, or run `galapagos serve` from a checkout of the repository to compile and serve the browser demo")
}
