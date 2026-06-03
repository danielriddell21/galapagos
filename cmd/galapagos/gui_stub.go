//go:build !ebiten

package main

import (
	"errors"
	"log/slog"
)

// launchGUI reports that this binary cannot open a native window. The native,
// OpenGL/Metal-backed window is built with the "ebiten" tag and is distributed
// only for macOS (via the Homebrew cask), where Metal needs no extra
// dependencies. On other platforms use `galapagos serve` for the browser demo,
// or pass --headless to train without a window.
func launchGUI(run guiRun, log *slog.Logger) error {
	return errors.New("native window unavailable in this build: on macOS install the Homebrew cask, on other platforms run `galapagos serve` for the browser demo, or pass --headless")
}
