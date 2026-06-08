package flappy

import (
	"image/color"

	"github.com/danielriddell21/galapagos/internal/core"
)

var (
	colPipe  = color.RGBA{60, 170, 90, 255}
	colBird  = color.RGBA{240, 210, 70, 255}
	colCrash = color.RGBA{220, 70, 60, 255}
	colEdge  = color.RGBA{90, 95, 110, 255}
)

// Bounds returns the world bounds for camera fitting.
func (e *Env) Bounds() (minX, minY, maxX, maxY float64) {
	return 0, 0, worldW, worldH
}

// Render draws the course: the ceiling and ground edges, each pipe's upper and
// lower bars, and the bird (red on a crash, yellow while alive).
func (e *Env) Render(r core.Renderer) {
	r.Line(0, 0, worldW, 0, colEdge)
	r.Line(0, worldH, worldW, worldH, colEdge)

	for _, p := range e.pipes {
		bar(r, p.x, 0, p.gapTop)                // upper bar
		bar(r, p.x, p.gapTop+gapHeight, worldH) // lower bar
	}

	c := color.Color(colBird)
	if e.done {
		c = colCrash
	}
	r.Circle(birdX, e.birdY, birdR, c)
}

// bar draws one filled pipe segment spanning the vertical range [y0, y1].
func bar(r core.Renderer, x, y0, y1 float64) {
	if y1 <= y0 {
		return
	}
	r.Polygon([][2]float64{{x, y0}, {x + pipeWidth, y0}, {x + pipeWidth, y1}, {x, y1}}, colPipe)
}
