package cube

import (
	"image/color"
	"math"
	"sort"

	rubix "github.com/danielriddell21/rubix/pkg/cube"
	"github.com/danielriddell21/rubix/pkg/render"

	"github.com/danielriddell21/galapagos/internal/core"
)

const (
	cell = 120.0
	NetW = cell
	NetH = cell
)

var cam = render.Camera{Yaw: 0.6, Pitch: 0.5, Scale: 22}

var bodyColor = color.RGBA{18, 18, 22, 255}

func (e *Env) Bounds() (minX, minY, maxX, maxY float64) { return 0, 0, NetW, NetH }

func (e *Env) Render(r core.Renderer) { RenderCube(r, e.c, rubix.Move(0), 0, 0, 0) }

type projQuad struct {
	pts   [4][2]float64
	col   color.Color
	depth float32
}

func RenderCube(r core.Renderer, c rubix.Cube, turn rubix.Move, frac float64, ox, oy float64) {
	cx, cy := ox+NetW/2, oy+NetH/2
	for _, q := range projectCube(c, turn, frac) {
		pts := make([][2]float64, len(q.pts))
		for i, p := range q.pts {
			pts[i] = [2]float64{cx + p[0], cy + p[1]}
		}
		r.Polygon(pts, q.col)
	}
}

func projectCube(c rubix.Cube, turn rubix.Move, frac float64) []projQuad {
	f := c.ToFacelets()
	sinY, cosY := sincos(cam.Yaw)
	sinX, cosX := sincos(cam.Pitch)

	animate := frac > 0 && turn < rubix.NumMoves
	var turnAxis int
	var tsin, tcos float32
	if animate {
		a, ang := render.TurnAngle(turn, clamp01(float32(frac)))
		turnAxis = a
		tsin, tcos = sincos(ang)
	}

	polys := render.BuildPolys()
	quads := make([]projQuad, 0, len(polys))
	for _, p := range polys {
		spin := animate && render.InTurnLayer(turn, p.Center)
		var pts [4][2]float64
		var depth float32
		for i := range 4 {
			pv := p.Cube[i]
			if spin {
				pv = render.RotateAxis(pv, turnAxis, tsin, tcos)
			}
			pv = render.RotateCamera(pv, sinY, cosY, sinX, cosX)
			pts[i] = [2]float64{float64(pv.X * cam.Scale), float64(-pv.Y * cam.Scale)}
			depth += pv.Z
		}
		col := color.Color(bodyColor)
		if p.Kind == render.Top {
			col = render.FaceletColor(f[int(p.Face)*9+p.Cell])
		}
		quads = append(quads, projQuad{pts: pts, col: col, depth: depth / 4})
	}
	sort.Slice(quads, func(i, j int) bool { return quads[i].depth < quads[j].depth })
	return quads
}

func sincos(a float32) (sin, cos float32) {
	s, c := math.Sincos(float64(a))
	return float32(s), float32(c)
}

func clamp01(t float32) float32 { return min(max(t, 0), 1) }
