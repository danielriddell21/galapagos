package flappy

import (
	"fmt"
	"math/rand/v2"

	"github.com/danielriddell21/galapagos/internal/core"
)

const (
	worldW = 100.0
	worldH = 100.0
	birdX  = 30.0
	birdR  = 3.0

	gravity     = 0.22
	flapImpulse = -2.4
	maxVY       = 5.0

	pipeSpeed  = 0.7
	pipeWidth  = 12.0
	gapHeight  = 40.0
	pipeSpace  = 50.0
	gapMargin  = 8.0
	firstPipeX = 70.0
	numPipes   = 3

	stepReward = 0.1
	clearBonus = 1.0
)

type Config struct {
	MaxSteps int
}

func DefaultConfig() Config { return Config{MaxSteps: 1000} }

type pipe struct {
	x      float64
	gapTop float64
	passed bool
}

type State struct{ obs []float64 }

func (s State) Observation() []float64 { return s.obs }

type Env struct {
	cfg     Config
	birdY   float64
	birdVY  float64
	pipes   []pipe
	cleared int
	t       int
	rng     *rand.Rand
	done    bool
}

func New(cfg Config) *Env { return &Env{cfg: cfg} }

func (e *Env) Reset(rng *rand.Rand) core.State {
	e.rng = rand.New(rand.NewPCG(rng.Uint64(), 0xf1a99))
	e.birdY = worldH / 2
	e.birdVY = 0
	e.cleared = 0
	e.t = 0
	e.done = false
	e.pipes = make([]pipe, numPipes)
	for i := range e.pipes {
		e.pipes[i] = pipe{x: firstPipeX + float64(i)*pipeSpace, gapTop: e.randGapTop()}
	}
	return e.state()
}

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
			e.cleared++
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

func (e *Env) rightmostX() float64 {
	m := e.pipes[0].x
	for _, p := range e.pipes[1:] {
		m = max(m, p.x)
	}
	return m
}

func (e *Env) randGapTop() float64 {
	return gapMargin + e.rng.Float64()*(worldH-gapHeight-2*gapMargin)
}

func (e *Env) Solved() bool { return !e.done }

func (e *Env) Status() string { return fmt.Sprintf("pipes %d", e.cleared) }

func (e *Env) ActionSpec() core.Spec {
	return core.Spec{Dim: 1, Low: []float64{0}, High: []float64{1}, Discrete: true}
}

func (e *Env) ObservationSpec() core.Spec {
	const n = 4
	high := make([]float64, n)
	for i := range high {
		high[i] = 1
	}
	return core.Spec{Dim: n, Low: make([]float64, n), High: high}
}

func actionIndex(a core.Action) int {
	v := a.Vector()
	if len(v) == 0 {
		return 0
	}
	return min(max(int(v[0]+0.5), 0), 1)
}

func clamp01(x float64) float64 { return min(max(x, 0), 1) }

var _ core.Environment = (*Env)(nil)
