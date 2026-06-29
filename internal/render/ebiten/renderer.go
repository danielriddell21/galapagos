//go:build ebiten

package ebiten

import (
	"image/color"

	eb "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/danielriddell21/galapagos/internal/core"
)

// Renderer draws core primitives onto an Ebiten image. World-space shapes
// (lines, circles, polygons) are transformed through the camera; text is drawn
// in screen space for the HUD. It is a thin forwarding shim: it holds no
// simulation state.
type Renderer struct {
	dst *eb.Image
	cam *core.Camera
}

// New returns a renderer with a camera sized to the given viewport in pixels.
func New(width, height int) *Renderer {
	return &Renderer{cam: core.NewCamera(float64(width), float64(height))}
}

// Begin sets the destination image for the current frame. Call once per frame
// before issuing draw calls.
func (r *Renderer) Begin(dst *eb.Image) { r.dst = dst }

// Camera implements core.Renderer.
func (r *Renderer) Camera() *core.Camera { return r.cam }

// Line implements core.Renderer, transforming endpoints through the camera.
func (r *Renderer) Line(x1, y1, x2, y2 float64, c color.Color) {
	sx1, sy1 := r.cam.WorldToScreen(x1, y1)
	sx2, sy2 := r.cam.WorldToScreen(x2, y2)
	vector.StrokeLine(r.dst, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 1, c, true)
}

// Circle implements core.Renderer, scaling the radius by the camera zoom.
func (r *Renderer) Circle(x, y, radius float64, c color.Color) {
	sx, sy := r.cam.WorldToScreen(x, y)
	vector.FillCircle(r.dst, float32(sx), float32(sy), float32(radius*r.cam.Zoom), c, true)
}

// Polygon implements core.Renderer by filling the closed shape through the
// transformed points.
func (r *Renderer) Polygon(pts [][2]float64, c color.Color) {
	if len(pts) < 3 {
		return
	}
	var path vector.Path
	sx, sy := r.cam.WorldToScreen(pts[0][0], pts[0][1])
	path.MoveTo(float32(sx), float32(sy))
	for _, p := range pts[1:] {
		x, y := r.cam.WorldToScreen(p[0], p[1])
		path.LineTo(float32(x), float32(y))
	}
	path.Close()

	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(r.dst, &path, nil, op)
}

// Text implements core.Renderer, drawing in screen space for HUD overlays.
func (r *Renderer) Text(x, y float64, s string) {
	ebitenutil.DebugPrintAt(r.dst, s, int(x), int(y))
}

var _ core.Renderer = (*Renderer)(nil)
