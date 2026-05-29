// Package render provides Renderer implementations. Nop is a headless,
// dependency-free renderer used for tests and turbo/batch training; the ebiten
// subpackage holds the windowed implementation. Both satisfy core.Renderer.
package render

import (
	"image/color"

	"github.com/danielriddell21/galapagos/internal/core"
)

// Nop is a renderer that discards all drawing calls. It lets the simulation run
// headlessly so all non-visual logic is testable without a graphics backend.
type Nop struct {
	cam *core.Camera
}

// NewNop returns a Nop renderer with a default off-screen camera.
func NewNop() *Nop {
	return &Nop{cam: core.NewCamera(1, 1)}
}

// Line implements core.Renderer.
func (n *Nop) Line(x1, y1, x2, y2 float64, c color.Color) {}

// Circle implements core.Renderer.
func (n *Nop) Circle(x, y, radius float64, c color.Color) {}

// Polygon implements core.Renderer.
func (n *Nop) Polygon(pts [][2]float64, c color.Color) {}

// Text implements core.Renderer.
func (n *Nop) Text(x, y float64, s string) {}

// Camera implements core.Renderer.
func (n *Nop) Camera() *core.Camera { return n.cam }

var _ core.Renderer = (*Nop)(nil)
