package core

import "github.com/danielriddell21/crucible/view"

// Camera is the family's display-free 2D pan/zoom camera (crucible/view),
// with follow and fit-to-bounds framing.
type Camera = view.Camera

// NewCamera returns a camera at unit zoom filling a viewW×viewH screen.
func NewCamera(viewW, viewH float64) *Camera { return view.New(viewW, viewH) }
