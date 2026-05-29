//go:build !ebiten

package main

import (
	"errors"

	"github.com/danielriddell21/galapagos/internal/config"
)

// launchGUI reports that this binary was built without the Ebiten window. The
// windowed demo requires a display and is built with the "ebiten" build tag;
// use --headless to train without a window.
func launchGUI(c config.Racing, out string) error {
	return errors.New("this binary was built without GUI support; rebuild with -tags ebiten on a machine with a display, or pass --headless")
}
