package gui

import (
	"github.com/danielriddell21/crucible/keymap"
	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/galapagos/internal/core"
)

type Config struct {
	Title      string
	Keymap     []keymap.Binding
	Population bool

	Step     func() bool
	Complete func()
	Next     func()
	Render   func(core.Renderer)
	HUD      func() []string
	Series   func() []float64

	Bounds  func() (minX, minY, maxX, maxY float64, ok bool)
	Leader  func() (x, y float64, ok bool)
	Sensors func() (origin core.Vec2, ends []core.Vec2, ok bool)

	Save       func() error
	Load       func() error
	Regenerate func(seed int64)

	// Rec names the recording [Render] writes; it is set by tools/demogen.
	Rec record.Options
}
