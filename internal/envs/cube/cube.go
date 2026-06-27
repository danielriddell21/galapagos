// Package cube wraps the external Rubik's cube model from
// github.com/danielriddell21/rubix/pkg/cube as a single-agent core.Environment.
// It reimplements no cube mechanics: stepping is state.Applied(move) and the
// terminal signal is cube.IsSolved. The state is the comparable cube value, so a
// tabular agent can key its Q-table on it directly.
package cube

import (
	"math/rand/v2"

	rubix "github.com/danielriddell21/rubix/pkg/cube"

	"github.com/danielriddell21/galapagos/internal/core"
)

// Config parameters an episode.
type Config struct {
	ScrambleDepth int // number of random moves applied to scramble the start
	MaxSteps      int // step budget before the episode ends unsolved
}

// DefaultConfig returns a shallow scramble and tight horizon that tabular
// Q-learning can reliably solve.
func DefaultConfig() Config { return Config{ScrambleDepth: 4, MaxSteps: 10} }

// State is a cube position. It holds only the comparable cube value, so it can
// be used directly as a Q-table key.
type State struct{ c rubix.Cube }

// Observation returns a normalized 54-facelet vector. It is computed on demand
// (no stored slice) so State stays comparable. Q-learning keys on the state
// itself and ignores this; it exists for the core.State contract.
func (s State) Observation() []float64 {
	f := s.c.ToFacelets()
	obs := make([]float64, len(f))
	for i, col := range f {
		obs[i] = float64(col) / 5
	}
	return obs
}

// Env is a Rubik's cube world.
type Env struct {
	cfg   Config
	c     rubix.Cube
	steps int
	done  bool
}

// New returns a cube environment.
func New(cfg Config) *Env { return &Env{cfg: cfg} }

// Reset scrambles a fresh cube. The scramble seed is drawn from rng, so a fixed
// reset stream (as sim.TrainAgent uses) trains on one reproducible scramble.
func (e *Env) Reset(rng *rand.Rand) core.State {
	e.c = rubix.ScrambledCube(e.cfg.ScrambleDepth, int64(rng.Uint64()))
	e.steps = 0
	e.done = false
	return State{e.c}
}

// Step applies the chosen move and reports the next state, reward, and done.
// Reward is a small per-step cost plus a terminal bonus for solving, so shorter
// solutions score higher.
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

// moveIndex maps an action vector to a move index in [0, NumMoves).
func moveIndex(a core.Action) int {
	v := a.Vector()
	if len(v) == 0 {
		return 0
	}
	return min(max(int(v[0]+0.5), 0), int(rubix.NumMoves)-1)
}

// Solved reports whether the cube is solved.
func (e *Env) Solved() bool { return e.c.IsSolved() }

// ActionSpec implements core.Environment: one discrete move in [0, NumMoves).
func (e *Env) ActionSpec() core.Spec {
	return core.Spec{Dim: 1, Low: []float64{0}, High: []float64{float64(rubix.NumMoves - 1)}, Discrete: true}
}

// ObservationSpec implements core.Environment.
func (e *Env) ObservationSpec() core.Spec {
	const n = 54
	high := make([]float64, n)
	for i := range high {
		high[i] = 1
	}
	return core.Spec{Dim: n, Low: make([]float64, n), High: high}
}

var _ core.Environment = (*Env)(nil)
