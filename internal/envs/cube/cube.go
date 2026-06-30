package cube

import (
	"math/rand/v2"

	rubix "github.com/danielriddell21/rubix/pkg/cube"

	"github.com/danielriddell21/galapagos/internal/core"
)

type Config struct {
	ScrambleDepth int
	MaxSteps      int
}

func DefaultConfig() Config { return Config{ScrambleDepth: 4, MaxSteps: 10} }

type State struct{ c rubix.Cube }

func (s State) Observation() []float64 {
	f := s.c.ToFacelets()
	obs := make([]float64, len(f))
	for i, col := range f {
		obs[i] = float64(col) / 5
	}
	return obs
}

type Env struct {
	cfg   Config
	c     rubix.Cube
	steps int
	done  bool
}

func New(cfg Config) *Env { return &Env{cfg: cfg} }

func (e *Env) Reset(rng *rand.Rand) core.State {
	e.c = rubix.ScrambledCube(e.cfg.ScrambleDepth, int64(rng.Uint64()))
	e.steps = 0
	e.done = false
	return State{e.c}
}

func (e *Env) Step(a core.Action) (core.State, core.Reward, bool) {
	if e.done {
		return State{e.c}, 0, true
	}
	e.c = e.c.Applied(rubix.Move(moveIndex(a)))
	e.steps++

	reward := core.Reward(-0.01)
	if e.c.IsSolved() {
		reward = 1
		e.done = true
	} else if e.steps >= e.cfg.MaxSteps {
		e.done = true
	}
	return State{e.c}, reward, e.done
}

func moveIndex(a core.Action) int {
	v := a.Vector()
	if len(v) == 0 {
		return 0
	}
	return min(max(int(v[0]+0.5), 0), int(rubix.NumMoves)-1)
}

func (e *Env) Solved() bool { return e.c.IsSolved() }

func (e *Env) ActionSpec() core.Spec {
	return core.Spec{Dim: 1, Low: []float64{0}, High: []float64{float64(rubix.NumMoves - 1)}, Discrete: true}
}

func (e *Env) ObservationSpec() core.Spec {
	const n = 54
	high := make([]float64, n)
	for i := range high {
		high[i] = 1
	}
	return core.Spec{Dim: n, Low: make([]float64, n), High: high}
}

var _ core.Environment = (*Env)(nil)
