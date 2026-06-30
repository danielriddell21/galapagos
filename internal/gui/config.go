// Package gui owns the Ebiten window and the uniform Run/Available seam. It
// drives a backend-agnostic Config — a bundle of closures supplied by the CLI —
// so the same window serves every environment and agent without the GUI
// importing any concrete environment, agent, or the simulation core directly.
//
// The Ebiten-backed implementation is compiled only under the "ebiten" build
// tag; without it Run is a stub reporting the GUI is unavailable, keeping the
// default build free of graphics and cgo dependencies.
package gui

import "github.com/danielriddell21/galapagos/internal/core"

// Config is a backend-agnostic description of a windowed run: a bundle of
// closures the Ebiten game drives, plus optional GIF-recording settings. It
// carries no Ebiten types, so the CLI builds it in either build and passes it
// to Run.
type Config struct {
	Title      string
	Keymap     []string
	Population bool // population run (generations) vs online run (episodes)

	Step     func() bool         // advance one tick; true when the gen/episode ends
	Complete func()              // called once when a gen/episode ends, before next
	Next     func()              // evolve / start the next episode
	Render   func(core.Renderer) // draw the world
	HUD      func() []string     // HUD lines, top-left
	Series   func() []float64    // fitness/return history for the sparkline

	Bounds  func() (minX, minY, maxX, maxY float64, ok bool)
	Leader  func() (x, y float64, ok bool)                       // follow target
	Sensors func() (origin core.Vec2, ends []core.Vec2, ok bool) // ray overlay

	Save       func() error
	Load       func() error
	Regenerate func(seed int64)

	// Recording: when RecordPath is set the window records a GIF and exits.
	RecordPath   string
	RecordFrames int
	RecordFPS    int
	RecordScale  int
}
