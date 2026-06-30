//go:build !ebiten

package gui

import (
	"errors"
	"log/slog"
)

// Available reports whether the Ebiten window is compiled in.
func Available() bool { return false }

// Run reports that this binary cannot open a native window. The native,
// OpenGL/Metal-backed window is built with the "ebiten" tag and is distributed
// only for macOS (via the Homebrew cask), where Metal needs no extra
// dependencies. On other platforms use `galapagos serve` for the browser demo,
// or pass --headless to train without a window.
func Run(run Config, log *slog.Logger) error {
	return errors.New("native window unavailable in this build: on macOS install the Homebrew cask, on other platforms run `galapagos serve` for the browser demo, or pass --headless")
}
