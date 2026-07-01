package chess

import (
	"math/rand/v2"

	gambit "github.com/danielriddell21/gambit/pkg/chess"

	"github.com/danielriddell21/galapagos/internal/core"
)

type Policy func(core.State) core.Action

type Config struct {
	MaxPlies int
	Opponent Policy
}

func DefaultConfig() Config { return Config{MaxPlies: 80} }

var pieceValue = [...]float64{0, 1, 3, 3, 5, 9, 0}

const shapeScale = 0.1

type State struct{ b *gambit.Board }

func (s State) Observation() []float64 { return encode(s.b) }

func (s State) Board() *gambit.Board { return s.b }

type Env struct {
	cfg   Config
	b     *gambit.Board
	last  gambit.Move
	plies int
	rng   *rand.Rand
	done  bool
}

func New(cfg Config) *Env {
	if cfg.MaxPlies <= 0 {
		cfg.MaxPlies = 80
	}
	return &Env{cfg: cfg}
}

func (e *Env) Reset(rng *rand.Rand) core.State {
	e.b = gambit.NewStartingBoard()
	e.rng = rand.New(rand.NewPCG(rng.Uint64(), 0xc4e55))
	e.last = gambit.NoMove
	e.plies = 0
	e.done = false
	return State{e.b}
}

func (e *Env) Step(a core.Action) (core.State, core.Reward, bool) {
	if e.done {
		return State{e.b}, 0, true
	}
	before := whiteAdvantage(e.b)

	mv, ok := bestLegalMove(e.b, a)
	if !ok { // White has no legal move: checkmate or stalemate
		e.done = true
		res, _ := e.b.Status()
		return State{e.b}, resultReward(res), true
	}
	e.apply(mv)
	if res, _ := e.b.Status(); res != gambit.InProgress {
		e.done = true
		return State{e.b}, resultReward(res), true
	}

	e.apply(e.opponentMove())
	if res, _ := e.b.Status(); res != gambit.InProgress {
		e.done = true
		return State{e.b}, resultReward(res), true
	}

	reward := core.Reward(shapeScale * (whiteAdvantage(e.b) - before))
	if e.plies >= e.cfg.MaxPlies {
		e.done = true
	}
	return State{e.b}, reward, e.done
}

func (e *Env) apply(m gambit.Move) {
	e.b = e.b.ApplyMove(m)
	e.last = m
	e.plies++
}

func (e *Env) opponentMove() gambit.Move {
	if e.cfg.Opponent != nil {
		if m, ok := bestLegalMove(e.b, e.cfg.Opponent(State{e.b})); ok {
			return m
		}
	}
	moves := e.b.LegalMoves()
	return moves[e.rng.IntN(len(moves))]
}

func resultReward(r gambit.Result) core.Reward {
	switch r {
	case gambit.WhiteWins:
		return 1
	case gambit.BlackWins:
		return -1
	default:
		return 0
	}
}

func whiteAdvantage(b *gambit.Board) float64 {
	var sum float64
	b.Each(func(_ gambit.Square, p gambit.Piece) {
		if p.IsEmpty() {
			return
		}
		v := pieceValue[p.Type()]
		if p.Color() == gambit.White {
			sum += v
		} else {
			sum -= v
		}
	})
	return sum
}

func encode(b *gambit.Board) []float64 {
	stm := b.SideToMove()
	obs := make([]float64, 64)
	b.Each(func(s gambit.Square, p gambit.Piece) {
		idx := perspective(s, stm)
		if p.IsEmpty() {
			obs[idx] = 0.5
			return
		}
		mag := float64(p.Type()) / 12 // pawn 1/12 .. king 6/12
		if p.Color() == stm {
			obs[idx] = 0.5 + mag
		} else {
			obs[idx] = 0.5 - mag
		}
	})
	return obs
}

func perspective(s gambit.Square, c gambit.Color) int {
	if c == gambit.White {
		return int(s)
	}
	return int(s) ^ 56
}

func (e *Env) ActionSpec() core.Spec {
	const n = 128
	return core.Spec{Dim: n, Low: make([]float64, n), High: ones(n)}
}

func (e *Env) ObservationSpec() core.Spec {
	const n = 64
	return core.Spec{Dim: n, Low: make([]float64, n), High: ones(n)}
}

func ones(n int) []float64 {
	v := make([]float64, n)
	for i := range v {
		v[i] = 1
	}
	return v
}

var _ core.Environment = (*Env)(nil)
