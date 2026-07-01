package chess

import (
	"image/color"

	gambit "github.com/danielriddell21/gambit/pkg/chess"

	"github.com/danielriddell21/galapagos/internal/core"
)

const cell = 60.0

var (
	colLight = color.RGBA{235, 215, 180, 255}
	colDark  = color.RGBA{150, 110, 75, 255}
	colMove  = color.RGBA{240, 220, 90, 255}
	colWhite = color.RGBA{70, 110, 180, 255}
	colBlack = color.RGBA{30, 30, 36, 255}
)

var pieceLetter = [...]byte{0: ' ', gambit.Pawn: 'P', gambit.Knight: 'N', gambit.Bishop: 'B', gambit.Rook: 'R', gambit.Queen: 'Q', gambit.King: 'K'}

func (e *Env) Bounds() (minX, minY, maxX, maxY float64) {
	return 0, 0, 8 * cell, 8 * cell
}

func (e *Env) Render(r core.Renderer) {
	for file := range 8 {
		for rank := range 8 {
			x0, y0 := squareXY(file, rank)
			bg := colDark
			if (file+rank)%2 == 1 {
				bg = colLight
			}
			if e.last != gambit.NoMove && (sameSquare(e.last.From(), file, rank) || sameSquare(e.last.To(), file, rank)) {
				bg = colMove
			}
			r.Polygon([][2]float64{{x0, y0}, {x0 + cell, y0}, {x0 + cell, y0 + cell}, {x0, y0 + cell}}, bg)
		}
	}

	// Pieces are discs tinted by colour; the type letter is drawn on top in screen
	// space (the Text primitive is not camera-transformed) via the camera.
	cam := r.Camera()
	e.b.Each(func(s gambit.Square, p gambit.Piece) {
		if p.IsEmpty() {
			return
		}
		x0, y0 := squareXY(s.File(), s.Rank())
		cx, cy := x0+cell/2, y0+cell/2
		disc := colWhite
		if p.Color() == gambit.Black {
			disc = colBlack
		}
		r.Circle(cx, cy, cell*0.34, disc)
		sx, sy := cam.WorldToScreen(cx, cy)
		r.Text(sx-3, sy-7, letterFor(p))
	})
}

func squareXY(file, rank int) (float64, float64) {
	return float64(file) * cell, float64(7-rank) * cell
}

func sameSquare(s gambit.Square, file, rank int) bool {
	return s.File() == file && s.Rank() == rank
}

func letterFor(p gambit.Piece) string {
	return string(rune(pieceLetter[p.Type()]))
}
