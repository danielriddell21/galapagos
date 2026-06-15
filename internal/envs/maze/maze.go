// Package maze implements a grid-maze environment as a single-agent
// core.Environment with discrete moves. It depends only on core, demonstrating
// that the agent interface generalizes across very different worlds.
package maze

import (
	"image/color"
	"math/rand/v2"

	"github.com/danielriddell21/galapagos/internal/core"
)

// Direction indices for actions and wall bitmasks.
const (
	north = iota
	east
	south
	west
)

var (
	dx = [4]int{0, 1, 0, -1}
	dy = [4]int{-1, 0, 1, 0}
)

// Config parameters the maze.
type Config struct {
	Width    int
	Height   int
	MaxSteps int
}

// DefaultConfig returns a small maze with a generous step budget.
func DefaultConfig() Config { return Config{Width: 8, Height: 8, MaxSteps: 200} }

// State is the maze observation: normalized agent position plus a wall flag in
// each direction.
type State struct{ obs []float64 }

// Observation implements core.State.
func (s State) Observation() []float64 { return s.obs }

// Env is a grid-maze world. open[c] holds which of the four directions are
// passable from cell c.
type Env struct {
	cfg   Config
	open  [][4]bool
	x, y  int
	steps int
	done  bool
}

// New returns a maze environment.
func New(cfg Config) *Env { return &Env{cfg: cfg} }

// cell returns the linear index of grid cell (x, y).
func (e *Env) cell(x, y int) int { return y*e.cfg.Width + x }

// Reset carves a new perfect maze with a depth-first backtracker and places the
// agent at the top-left corner.
func (e *Env) Reset(rng *rand.Rand) core.State {
	w, h := e.cfg.Width, e.cfg.Height
	e.open = make([][4]bool, w*h)
	visited := make([]bool, w*h)
	type pos struct{ x, y int }
	stack := []pos{{0, 0}}
	visited[e.cell(0, 0)] = true
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		var dirs []int
		for d := range 4 {
			nx, ny := cur.x+dx[d], cur.y+dy[d]
			if nx >= 0 && nx < w && ny >= 0 && ny < h && !visited[e.cell(nx, ny)] {
				dirs = append(dirs, d)
			}
		}
		if len(dirs) == 0 {
			stack = stack[:len(stack)-1]
			continue
		}
		d := dirs[rng.IntN(len(dirs))]
		nx, ny := cur.x+dx[d], cur.y+dy[d]
		e.open[e.cell(cur.x, cur.y)][d] = true
		e.open[e.cell(nx, ny)][(d+2)%4] = true
		visited[e.cell(nx, ny)] = true
		stack = append(stack, pos{nx, ny})
	}
	e.x, e.y, e.steps, e.done = 0, 0, 0, false
	return e.observe()
}

// Step moves the agent in the chosen direction if no wall blocks it.
func (e *Env) Step(a core.Action) (core.State, core.Reward, bool) {
	if e.done {
		return e.observe(), 0, true
	}
	d := direction(a)
	if e.open[e.cell(e.x, e.y)][d] {
		e.x += dx[d]
		e.y += dy[d]
	}
	e.steps++

	reward := core.Reward(-0.01) // small per-step cost encourages short paths
	if e.x == e.cfg.Width-1 && e.y == e.cfg.Height-1 {
		reward = 1
		e.done = true
	} else if e.steps >= e.cfg.MaxSteps {
		e.done = true
	}
	return e.observe(), reward, e.done
}

// observe reports normalized position and the open/blocked state of each wall.
func (e *Env) observe() core.State {
	w := e.open[e.cell(e.x, e.y)]
	wall := func(d int) float64 {
		if w[d] {
			return 1
		}
		return 0
	}
	return State{obs: []float64{
		float64(e.x) / float64(e.cfg.Width-1),
		float64(e.y) / float64(e.cfg.Height-1),
		wall(north), wall(east), wall(south), wall(west),
	}}
}

// ActionSpec implements core.Environment: a discrete direction in [0,3].
func (e *Env) ActionSpec() core.Spec {
	return core.Spec{Dim: 1, Low: []float64{0}, High: []float64{3}, Discrete: true}
}

// ObservationSpec implements core.Environment.
func (e *Env) ObservationSpec() core.Spec {
	return core.Spec{Dim: 6, Low: make([]float64, 6), High: []float64{1, 1, 1, 1, 1, 1}}
}

// Solved reports whether the agent has reached the goal.
func (e *Env) Solved() bool { return e.x == e.cfg.Width-1 && e.y == e.cfg.Height-1 }

// Bounds returns the world bounds of the maze drawing for camera fitting. It
// matches the cell size used by Render.
func (e *Env) Bounds() (minX, minY, maxX, maxY float64) {
	const s = 32.0
	return 0, 0, float64(e.cfg.Width) * s, float64(e.cfg.Height) * s
}

var _ core.Environment = (*Env)(nil)

// direction maps an action vector to a direction index in [0,3].
func direction(a core.Action) int {
	v := a.Vector()
	if len(v) == 0 {
		return 0
	}
	d := int(v[0] + 0.5)
	return min(max(d, 0), 3)
}

// Render draws the goal cell, the maze walls, and the agent. The goal and the
// agent light up green once the exit is reached.
func (e *Env) Render(r core.Renderer) {
	const s = 32.0
	const pad = 5.0
	solved := e.Solved()

	// Goal marker at the bottom-right exit cell, brighter once reached.
	goalCol := color.RGBA{60, 140, 70, 255}
	if solved {
		goalCol = color.RGBA{90, 230, 110, 255}
	}
	gx, gy := float64(e.cfg.Width-1)*s, float64(e.cfg.Height-1)*s
	r.Polygon([][2]float64{{gx + pad, gy + pad}, {gx + s - pad, gy + pad}, {gx + s - pad, gy + s - pad}, {gx + pad, gy + s - pad}}, goalCol)

	for y := range e.cfg.Height {
		for x := range e.cfg.Width {
			w := e.open[e.cell(x, y)]
			x0, y0 := float64(x)*s, float64(y)*s
			wallColor := color.RGBA{120, 120, 130, 255}
			if !w[north] {
				r.Line(x0, y0, x0+s, y0, wallColor)
			}
			if !w[west] {
				r.Line(x0, y0, x0, y0+s, wallColor)
			}
			if !w[east] {
				r.Line(x0+s, y0, x0+s, y0+s, wallColor)
			}
			if !w[south] {
				r.Line(x0, y0+s, x0+s, y0+s, wallColor)
			}
		}
	}
	agentCol := color.RGBA{80, 180, 255, 255}
	if solved {
		agentCol = color.RGBA{90, 230, 110, 255}
	}
	r.Circle(float64(e.x)*s+s/2, float64(e.y)*s+s/2, s/3, agentCol)
}

// Status reports whether the agent has reached the goal, for the HUD.
func (e *Env) Status() string {
	if e.Solved() {
		return "solved"
	}
	return "exploring"
}
