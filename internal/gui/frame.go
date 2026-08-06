package gui

import (
	"fmt"
	"image/color"
	"time"

	"github.com/danielriddell21/galapagos/internal/core"
)

// The visualizer's fixed logical frame size. Windows are resizable; the frame
// is scaled to fit.
const (
	screenW = 1024
	screenH = 768
)

// Background is the colour behind every frame. The caller clears with it,
// since how you clear differs between a live window and a pixel buffer.
var Background = color.RGBA{R: 18, G: 20, B: 26, A: 255}

// FrameState is the view state that belongs to the run rather than the world:
// how fast it is going, whether it is paused, and what extras are shown.
type FrameState struct {
	Speed    int
	Paused   bool
	Elapsed  time.Duration
	ShowRays bool
	Follow   bool
}

// DrawFrame composes one frame — the world, the sensor rays, the HUD, the
// keymap and the fitness sparkline — through r. It draws only through
// [core.Renderer], so the same code composes a live window frame and a
// headless one for documentation media. The caller clears to [Background]
// first.
func DrawFrame(r core.Renderer, cfg Config, st FrameState) {
	aimCamera(r, cfg, st.Follow)
	cfg.Render(r)
	if st.ShowRays {
		if origin, ends, ok := cfg.Sensors(); ok {
			for _, e := range ends {
				r.Line(origin.X, origin.Y, e.X, e.Y, color.RGBA{G: 200, B: 120, A: 160})
			}
		}
	}
	drawHUD(r, cfg, st)
	drawKeymap(r, cfg)
	drawSparkline(r, cfg.Series())
}

// aimCamera follows the leader when asked, and otherwise frames the whole
// scene.
func aimCamera(r core.Renderer, cfg Config, follow bool) {
	cam := r.Camera()
	if follow {
		if x, y, ok := cfg.Leader(); ok {
			cam.Zoom = 1.2
			cam.Follow(x, y)
			return
		}
	}
	if minX, minY, maxX, maxY, ok := cfg.Bounds(); ok {
		cam.FitBounds(minX, minY, maxX, maxY, 0.15)
	}
}

func drawHUD(r core.Renderer, cfg Config, st FrameState) {
	y := 10
	for _, line := range cfg.HUD() {
		r.Text(10, float64(y), line)
		y += 16
	}
	r.Text(10, float64(y), fmt.Sprintf("speed x%d%s", st.Speed, pausedLabel(st.Paused)))
	r.Text(10, float64(y+16), fmt.Sprintf("elapsed %s", st.Elapsed.Round(time.Second)))
}

func drawKeymap(r core.Renderer, cfg Config) {
	y := 10
	for _, line := range cfg.Keymap {
		r.Text(screenW-150, float64(y), line)
		y += 16
	}
}

func drawSparkline(r core.Renderer, series []float64) {
	if len(series) < 2 {
		return
	}
	const x0, y0, w, h = 10.0, 700.0, 240.0, 50.0
	lo, hi := series[0], series[0]
	for _, v := range series {
		lo, hi = min(lo, v), max(hi, v)
	}
	span := max(hi-lo, 1e-9)

	cam := r.Camera()
	prev := *cam
	// Centering the camera on the viewport with unit zoom makes world == screen
	// coordinates, so the HUD draws in screen space.
	*cam = core.Camera{X: screenW / 2, Y: screenH / 2, Zoom: 1, ViewW: screenW, ViewH: screenH}
	for i := 1; i < len(series); i++ {
		ax := x0 + w*float64(i-1)/float64(len(series)-1)
		bx := x0 + w*float64(i)/float64(len(series)-1)
		ay := y0 + h - h*(series[i-1]-lo)/span
		by := y0 + h - h*(series[i]-lo)/span
		r.Line(ax, ay, bx, by, color.RGBA{R: 255, G: 215, A: 255})
	}
	*cam = prev
}

func pausedLabel(p bool) string {
	if p {
		return " (paused)"
	}
	return ""
}
