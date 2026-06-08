// Package flappy is a side-scrolling "flap through the gaps" environment. A bird
// at a fixed horizontal position falls under gravity and flaps to rise, weaving
// through scrolling pipe gaps. The two-action discrete control and continuous
// state make it a natural fit for a function-approximation agent (DQN). It
// reimplements no external mechanics and imports only core and the standard
// library.
package flappy

import (
	"math/rand/v2"

	"github.com/danielriddell21/galapagos/internal/core"
)

// World geometry and dynamics, in abstract world units (the camera fits them).
const (
	worldW = 100.0 // visible width
	worldH = 100.0 // height (y grows downward; 0 is the ceiling)
	birdX  = 30.0  // the bird's fixed horizontal position
	birdR  = 3.0   // bird radius (for collisions and drawing)

	gravity     = 0.22 // added to vertical velocity each step
	flapImpulse = -2.4 // velocity set on a flap (upward is negative)
	maxVY       = 5.0  // velocity magnitude used to normalize the observation

	pipeSpeed  = 0.7  // leftward scroll per step
	pipeWidth  = 12.0 // horizontal thickness of a pipe
	gapHeight  = 40.0 // vertical opening in a pipe
	pipeSpace  = 50.0 // horizontal spacing between pipes
	gapMargin  = 8.0  // keep gaps away from the ceiling and ground
	firstPipeX = 70.0 // x of the first pipe at reset
	numPipes   = 3    // pipes tracked at once (recycled as they exit left)

	stepReward = 0.1 // reward per surviving step
	clearBonus = 1.0 // reward for clearing a pipe
)

// Config parameters an episode.
type Config struct {
	MaxSteps int // step budget before the episode ends
}

// DefaultConfig returns a generous horizon.
func DefaultConfig() Config { return Config{MaxSteps: 1000} }

// pipe is one obstacle: a vertical gap of height gapHeight whose top is at gapTop,
// with solid bars above and below. passed records that the bird has cleared it.
type pipe struct {
	x      float64
	gapTop float64
	passed bool
}

// State is the bird's observation. It carries the feature slice directly; the
// agent learns from it via function approximation (no comparable-state keying).
type State struct{ obs []float64 }

// Observation implements core.State.
func (s State) Observation() []float64 { return s.obs }

// Env is a flappy world.
type Env struct {
	cfg    Config
	birdY  float64
	birdVY float64
	pipes  []pipe
	t      int
	rng    *rand.Rand
	done   bool
}

// New returns a flappy environment.
func New(cfg Config) *Env { return &Env{cfg: cfg} }

// Reset centers the bird and lays out a fresh course. The pipe layout is seeded
// from rng, so a fixed reset stream trains on one reproducible course.
func (e *Env) Reset(rng *rand.Rand) core.State {
	e.rng = rand.New(rand.NewPCG(rng.Uint64(), 0xf1a99))
	e.birdY = worldH / 2
	e.birdVY = 0
	e.t = 0
	e.done = false
	e.pipes = make([]pipe, numPipes)
	for i := range e.pipes {
		e.pipes[i] = pipe{x: firstPipeX + float64(i)*pipeSpace, gapTop: e.randGapTop()}
	}
	return e.state()
}

// Step applies the action (1 flaps, anything else does nothing), advances the
// physics, scrolls the course, and reports the next state, reward, and done.
func (e *Env) Step(a core.Action) (core.State, core.Reward, bool) {
	if e.done {
		return e.state(), 0, true
	}
	if actionIndex(a) == 1 {
		e.birdVY = flapImpulse
	}
	e.birdVY += gravity
	e.birdY += e.birdVY

	reward := core.Reward(stepReward)
	for i := range e.pipes {
		e.pipes[i].x -= pipeSpeed
		p := &e.pipes[i]
		if !p.passed && p.x+pipeWidth < birdX {
			p.passed = true
			reward += clearBonus
		}
		if p.x+pipeWidth < 0 {
			p.x = e.rightmostX() + pipeSpace
			p.gapTop = e.randGapTop()
			p.passed = false
		}
	}
	e.t++

	if e.crashed() {
		e.done = true
		return e.state(), 0, true
	}
	if e.t >= e.cfg.MaxSteps {
		e.done = true
	}
	return e.state(), reward, e.done
}

// crashed reports whether the bird has hit the ceiling, the ground, or a pipe.
func (e *Env) crashed() bool {
	if e.birdY-birdR < 0 || e.birdY+birdR > worldH {
		return true
	}
	for _, p := range e.pipes {
		if birdX+birdR > p.x && birdX-birdR < p.x+pipeWidth {
			if e.birdY-birdR < p.gapTop || e.birdY+birdR > p.gapTop+gapHeight {
				return true
			}
		}
	}
	return false
}

// state builds the normalized observation: bird height, bird velocity, distance
// to the next pipe, and the next gap's center relative to the bird, each in [0,1].
func (e *Env) state() core.State {
	np := e.nextPipe()
	gapCenter := np.gapTop + gapHeight/2
	return State{obs: []float64{
		clamp01(e.birdY / worldH),
		clamp01((e.birdVY + maxVY) / (2 * maxVY)),
		clamp01((np.x - birdX) / worldW),
		clamp01((gapCenter - e.birdY + worldH) / (2 * worldH)),
	}}
}

// nextPipe returns the nearest pipe the bird has not yet cleared, or the
// rightmost pipe if all are behind it.
func (e *Env) nextPipe() pipe {
	best := -1
	for i, p := range e.pipes {
		if p.x+pipeWidth >= birdX && (best == -1 || p.x < e.pipes[best].x) {
			best = i
		}
	}
	if best == -1 {
		for i, p := range e.pipes {
			if best == -1 || p.x > e.pipes[best].x {
				best = i
			}
		}
	}
	return e.pipes[best]
}

// rightmostX returns the largest pipe x currently in play.
func (e *Env) rightmostX() float64 {
	m := e.pipes[0].x
	for _, p := range e.pipes[1:] {
		m = max(m, p.x)
	}
	return m
}

// randGapTop draws a gap position that keeps the opening clear of the edges.
func (e *Env) randGapTop() float64 {
	return gapMargin + e.rng.Float64()*(worldH-gapHeight-2*gapMargin)
}

// Solved reports whether the bird is currently alive (used for rendering accent).
func (e *Env) Solved() bool { return !e.done }

// ActionSpec implements core.Environment: one discrete action in {0 idle, 1 flap}.
func (e *Env) ActionSpec() core.Spec {
	return core.Spec{Dim: 1, Low: []float64{0}, High: []float64{1}, Discrete: true}
}

// ObservationSpec implements core.Environment.
func (e *Env) ObservationSpec() core.Spec {
	const n = 4
	high := make([]float64, n)
	for i := range high {
		high[i] = 1
	}
	return core.Spec{Dim: n, Low: make([]float64, n), High: high}
}

// actionIndex decodes the discrete action from an action vector.
func actionIndex(a core.Action) int {
	v := a.Vector()
	if len(v) == 0 {
		return 0
	}
	return min(max(int(v[0]+0.5), 0), 1)
}

// clamp01 clamps x to [0, 1].
func clamp01(x float64) float64 { return min(max(x, 0), 1) }

var _ core.Environment = (*Env)(nil)
