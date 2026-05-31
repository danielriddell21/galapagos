package cube

import (
	"image/color"

	"github.com/danielriddell21/galapagos/internal/core"
	rubix "github.com/danielriddell21/rubix/pkg/cube"
)

// cell is the world size of one facelet square.
const cell = 24.0

// facePos gives the top-left grid position (in cells) of each face in the
// unfolded net, indexed in facelet order U, R, F, D, L, B:
//
//	   U
//	L  F  R  B
//	   D
var facePos = [6][2]int{
	{3, 0}, // U
	{6, 3}, // R
	{3, 3}, // F
	{3, 6}, // D
	{0, 3}, // L
	{9, 3}, // B
}

// faceColors is the default sticker scheme: U white, R red, F green, D yellow,
// L orange, B blue.
var faceColors = [6]color.RGBA{
	{245, 245, 245, 255},
	{200, 40, 40, 255},
	{40, 170, 70, 255},
	{240, 220, 40, 255},
	{240, 150, 40, 255},
	{40, 90, 200, 255},
}

// Bounds returns the world bounds of the unfolded net for camera fitting.
func (e *Env) Bounds() (minX, minY, maxX, maxY float64) {
	return 0, 0, 12 * cell, 9 * cell
}

// Render draws the cube as an unfolded net of 54 filled facelets with borders.
func (e *Env) Render(r core.Renderer) {
	border := color.RGBA{20, 20, 24, 255}
	f := e.c.ToFacelets()
	for face := range 6 {
		baseCol, baseRow := facePos[face][0], facePos[face][1]
		for row := range 3 {
			for col := range 3 {
				sticker := f[face*9+row*3+col]
				x0 := float64(baseCol+col) * cell
				y0 := float64(baseRow+row) * cell
				pts := [][2]float64{{x0, y0}, {x0 + cell, y0}, {x0 + cell, y0 + cell}, {x0, y0 + cell}}
				r.Polygon(pts, faceColors[sticker])
				drawBorder(r, x0, y0, border)
			}
		}
	}
}

// drawBorder outlines one facelet cell.
func drawBorder(r core.Renderer, x0, y0 float64, c color.Color) {
	r.Line(x0, y0, x0+cell, y0, c)
	r.Line(x0+cell, y0, x0+cell, y0+cell, c)
	r.Line(x0+cell, y0+cell, x0, y0+cell, c)
	r.Line(x0, y0+cell, x0, y0, c)
}

// ensure the rubix color count assumption (6 faces) stays in sync.
var _ = [1]struct{}{}[6-int(rubix.ColB+1)]
