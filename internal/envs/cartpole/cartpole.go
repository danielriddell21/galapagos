// Package cartpole implements the classic cart-pole balancing task as a
// single-agent core.Environment. It depends only on core, so any agent — the
// genetic algorithm included — can balance the pole without modification.
package cartpole

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/danielriddell21/galapagos/internal/core"
)

// Physical constants of the classic cart-pole, matching the common formulation.
const (
	gravity      = 9.8
	massCart     = 1.0
	massPole     = 0.1
	totalMass    = massCart + massPole
	halfPole     = 0.5 // half the pole's length
	poleMassLen  = massPole * halfPole
	forceMag     = 10.0
	tau          = 0.02 // seconds between updates
	xThreshold   = 2.4
	angThreshold = 12 * math.Pi / 180
)

// Config parameters an episode.
type Config struct {
	MaxSteps int
}

// DefaultConfig returns the standard 500-step episode budget.
func DefaultConfig() Config { return Config{MaxSteps: 500} }

// State is the cart-pole observation: cart position and velocity, pole angle and
// angular velocity, each normalized to [0,1].
type State struct{ obs []float64 }

// Observation implements core.State.
func (s State) Observation() []float64 { return s.obs }

// Env is a cart-pole world.
type Env struct {
	cfg                Config
	x, xDot, th, thDot float64
	steps              int
	done               bool
}

// New returns a cart-pole environment.
func New(cfg Config) *Env { return &Env{cfg: cfg} }

// Reset starts a new episode with a small random perturbation.
func (e *Env) Reset(rng *rand.Rand) core.State {
	r := func() float64 { return (rng.Float64()*2 - 1) * 0.05 }
	e.x, e.xDot, e.th, e.thDot = r(), r(), r(), r()
	e.steps = 0
	e.done = false
	return e.observe()
}

// Step applies a left/right force based on the sign of the action and
// integrates the dynamics one timestep.
func (e *Env) Step(a core.Action) (core.State, core.Reward, bool) {
	if e.done {
		return e.observe(), 0, true
	}
	force := -forceMag
	if v := a.Vector(); len(v) > 0 && v[0] > 0 {
		force = forceMag
	}

	cosTh, sinTh := math.Cos(e.th), math.Sin(e.th)
	temp := (force + poleMassLen*e.thDot*e.thDot*sinTh) / totalMass
	thAcc := (gravity*sinTh - cosTh*temp) / (halfPole * (4.0/3.0 - massPole*cosTh*cosTh/totalMass))
	xAcc := temp - poleMassLen*thAcc*cosTh/totalMass

	e.x += tau * e.xDot
	e.xDot += tau * xAcc
	e.th += tau * e.thDot
	e.thDot += tau * thAcc
	e.steps++

	failed := math.Abs(e.x) > xThreshold || math.Abs(e.th) > angThreshold
	e.done = failed || e.steps >= e.cfg.MaxSteps
	return e.observe(), 1, e.done
}

// observe normalizes the raw state into [0,1].
func (e *Env) observe() core.State {
	norm := func(v, lo, hi float64) float64 {
		return min(max((v-lo)/(hi-lo), 0), 1)
	}
	return State{obs: []float64{
		norm(e.x, -xThreshold, xThreshold),
		(math.Tanh(e.xDot) + 1) / 2,
		norm(e.th, -angThreshold, angThreshold),
		(math.Tanh(e.thDot) + 1) / 2,
	}}
}

// ActionSpec implements core.Environment: one value selecting the push
// direction. The discrete bounds [0,1] mean two actions (left, right); Step
// pushes right when the value is positive.
func (e *Env) ActionSpec() core.Spec {
	return core.Spec{Dim: 1, Low: []float64{0}, High: []float64{1}, Discrete: true}
}

// Bounds returns the world bounds of the cart-and-rail drawing for camera
// fitting.
func (e *Env) Bounds() (minX, minY, maxX, maxY float64) {
	const scale = 100
	return -xThreshold * scale, -1.2 * scale, xThreshold * scale, 0.4 * scale
}

// ObservationSpec implements core.Environment.
func (e *Env) ObservationSpec() core.Spec {
	return core.Spec{Dim: 4, Low: []float64{0, 0, 0, 0}, High: []float64{1, 1, 1, 1}}
}

var _ core.Environment = (*Env)(nil)

// Render draws the cart and pole, scaled to a notional world width.
func (e *Env) Render(r core.Renderer) {
	const scale = 100
	cartY := 0.0
	cx := e.x * scale
	r.Line(-xThreshold*scale, cartY, xThreshold*scale, cartY, color.RGBA{120, 120, 130, 255})
	r.Circle(cx, cartY, 8, color.RGBA{80, 180, 255, 255})
	tipX := cx + math.Sin(e.th)*halfPole*2*scale
	tipY := cartY - math.Cos(e.th)*halfPole*2*scale
	r.Line(cx, cartY, tipX, tipY, color.RGBA{255, 215, 0, 255})
}
