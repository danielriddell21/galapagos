// Package soft renders galapagos frames into a software pixel buffer, with no
// display and no Ebiten. It implements the same [core.Renderer] the windowed
// renderer does, so the identical drawing code produces either a live frame or
// documentation media.
package soft

import (
	"image"
	"image/color"

	"github.com/danielriddell21/crucible/canvas"

	"github.com/danielriddell21/galapagos/internal/core"
)

// textAscent lifts a top-left text origin onto the built-in face's baseline,
// so text lands where the windowed renderer's debug printer puts it.
const textAscent = 11

// Renderer draws through a software canvas.
type Renderer struct {
	c   *canvas.Canvas
	cam *core.Camera
	w   int
	h   int
}

// New returns a renderer drawing into a width x height pixel buffer.
func New(width, height int) *Renderer {
	return &Renderer{
		c:   canvas.New(width, height),
		cam: core.NewCamera(float64(width), float64(height)),
		w:   width,
		h:   height,
	}
}

// Camera returns the view transform shared with the windowed renderer.
func (r *Renderer) Camera() *core.Camera { return r.cam }

// Clear paints the whole frame, ready for the next one.
func (r *Renderer) Clear(col color.RGBA) { r.c.Fill(col) }

// Image returns the frame just drawn. The pixels are reused by the next
// Clear, so a caller that keeps the frame must copy it.
func (r *Renderer) Image() *image.RGBA {
	return &image.RGBA{Pix: r.c.Pixels(), Stride: r.w * 4, Rect: image.Rect(0, 0, r.w, r.h)}
}

// Line strokes a one-pixel world-space line.
func (r *Renderer) Line(x1, y1, x2, y2 float64, c color.Color) {
	sx1, sy1 := r.cam.WorldToScreen(x1, y1)
	sx2, sy2 := r.cam.WorldToScreen(x2, y2)
	r.c.Line(sx1, sy1, sx2, sy2, 1, rgba(c))
}

// Circle fills a world-space disc, scaled by the camera's zoom.
func (r *Renderer) Circle(x, y, radius float64, c color.Color) {
	sx, sy := r.cam.WorldToScreen(x, y)
	r.c.Circle(sx, sy, radius*r.cam.Zoom, rgba(c))
}

// Polygon fills a closed world-space polygon.
func (r *Renderer) Polygon(pts [][2]float64, c color.Color) {
	if len(pts) < 3 {
		return
	}
	screen := make([][2]float64, len(pts))
	for i, p := range pts {
		x, y := r.cam.WorldToScreen(p[0], p[1])
		screen[i] = [2]float64{x, y}
	}
	r.c.Polygon(screen, rgba(c))
}

// Text draws a screen-space label from its top-left corner, in white, as the
// windowed renderer's debug printer does.
func (r *Renderer) Text(x, y float64, s string) {
	r.c.Text(int(x), int(y)+textAscent, s, color.RGBA{R: 255, G: 255, B: 255, A: 255})
}

// rgba converts a drawing colour to the concrete type the canvas takes.
func rgba(c color.Color) color.RGBA {
	if v, ok := c.(color.RGBA); ok {
		return v
	}
	v, _ := color.RGBAModel.Convert(c).(color.RGBA)
	return v
}

var _ core.Renderer = (*Renderer)(nil)
