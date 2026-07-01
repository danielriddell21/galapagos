package render

import (
	"image/color"

	"github.com/danielriddell21/galapagos/internal/core"
)

type Nop struct {
	cam *core.Camera
}

func NewNop() *Nop {
	return &Nop{cam: core.NewCamera(1, 1)}
}

func (n *Nop) Line(x1, y1, x2, y2 float64, c color.Color) {}

func (n *Nop) Circle(x, y, radius float64, c color.Color) {}

func (n *Nop) Polygon(pts [][2]float64, c color.Color) {}

func (n *Nop) Text(x, y float64, s string) {}

func (n *Nop) Camera() *core.Camera { return n.cam }

var _ core.Renderer = (*Nop)(nil)
