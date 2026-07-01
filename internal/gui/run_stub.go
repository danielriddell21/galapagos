//go:build !ebiten

package gui

import (
	"errors"
	"log/slog"
)

func Available() bool { return false }

func Run(run Config, log *slog.Logger) error {
	return errors.New("native window unavailable in this build: on macOS install the Homebrew cask, on other platforms run `galapagos serve` for the browser demo, or pass --headless")
}
