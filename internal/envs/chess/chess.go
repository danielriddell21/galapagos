// Package chess wraps the external chess engine from
// github.com/danielriddell21/gambit/pkg/chess as a single-agent
// core.Environment. It reimplements no chess mechanics: move generation,
// legality, and result detection all come from gambit. The learner plays White
// and the environment replies as Black with a configurable opponent (a random
// legal mover, or any injected policy — used for hall-of-fame co-evolution), so
// a two-player game fits the single-agent contract like tic-tac-toe.
//
// Boards are encoded from the side-to-move's perspective (vertically flipped for
// Black), so one evolved network can play either colour — which is what lets a
// member become the opponent in co-evolution.
package chess

import (
	"math/rand/v2"

	gambit "github.com/danielriddell21/gambit/pkg/chess"

	"github.com/danielriddell21/galapagos/internal/core"
)

// Policy chooses a move for a position, returning a preference vector the
// environment decodes against the legal moves. It is a plain function, not an
// agent, so the environment stays decoupled from the agents package.
type Policy func(core.State) core.Action

// Config parameters an episode.
type Config struct {
	MaxPlies int    // ply budget before the game is called a draw
	Opponent Policy // Black's policy; nil plays uniform-random legal moves
}

// DefaultConfig plays a random opponent over a bounded game.
func DefaultConfig() Config { return Config{MaxPlies: 80} }

// pieceValue is centipawn-free material by piece type (index by gambit.PieceType
// 0..6: none, pawn, knight, bishop, rook, queen, king).
var pieceValue = [...]float64{0, 1, 3, 3, 5, 9, 0}

// shapeScale weights the per-ply material-delta reward so captures give a dense
// signal while the terminal win/loss (±1) still dominates.
const shapeScale = 0.1

// State is a chess position. Observation is computed on demand from the side to
// move's perspective.
type State struct{ b *gambit.Board }

// Observation implements core.State: 64 squares from the side-to-move's
// perspective, own pieces above 0.5 and enemy pieces below, by material rank.
func (s State) Observation() []float64 { return encode(s.b) }

// Board returns the underlying gambit position. It is exported so a search-based
// agent (evochess) can generate and evaluate moves directly; network learners
// use only Observation. Applying a move returns a new board, so the returned
// pointer is safe to read.
func (s State) Board() *gambit.Board { return s.b }

// Env is a chess world. The learner is White; Black is the configured opponent.
type Env struct {
	cfg   Config
	b     *gambit.Board
	last  gambit.Move
	plies int
	rng   *rand.Rand
	done  bool
}

// New returns a chess environment.
func New(cfg Config) *Env {
	if cfg.MaxPlies <= 0 {
		cfg.MaxPlies = 80
	}
	return &Env{cfg: cfg}
}

// Reset starts a fresh game from the standard position and seeds the random
// opponent from rng, so a fixed reset stream trains against one reproducible
// opponent stream.
func (e *Env) Reset(rng *rand.Rand) core.State {
	e.b = gambit.NewStartingBoard()
	e.rng = rand.New(rand.NewPCG(rng.Uint64(), 0xc4e55))
	e.last = gambit.NoMove
	e.plies = 0
	e.done = false
	return State{e.b}
}

// Step plays the learner's (White's) move, then the opponent's (Black's) reply,
// and reports the next state, reward, and whether the game is over. Terminal
// games score win +1 / draw 0 / loss −1; otherwise the reward is the change in
// White's material advantage across the ply pair, so capturing is rewarded.
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

// apply makes a move, advancing the board and ply count and recording the move
// for rendering.
func (e *Env) apply(m gambit.Move) {
	e.b = e.b.ApplyMove(m)
	e.last = m
	e.plies++
}

// opponentMove picks Black's reply: the injected policy's best legal move, or a
// uniform-random legal move when no policy is set.
func (e *Env) opponentMove() gambit.Move {
	if e.cfg.Opponent != nil {
		if m, ok := bestLegalMove(e.b, e.cfg.Opponent(State{e.b})); ok {
			return m
		}
	}
	moves := e.b.LegalMoves()
	return moves[e.rng.IntN(len(moves))]
}

// resultReward maps a finished game to White's reward.
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

// whiteAdvantage is White material minus Black material on the board.
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

// encode returns the 64-square observation from the side-to-move's perspective.
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

// perspective maps a board square to an index from c's point of view: identity
// for White, vertically flipped (rank-mirrored) for Black, so the side to move
// always looks "up the board."
func perspective(s gambit.Square, c gambit.Color) int {
	if c == gambit.White {
		return int(s)
	}
	return int(s) ^ 56
}

// ActionSpec implements core.Environment: 64 "from" preferences followed by 64
// "to" preferences, decoded against the legal moves.
func (e *Env) ActionSpec() core.Spec {
	const n = 128
	return core.Spec{Dim: n, Low: make([]float64, n), High: ones(n)}
}

// ObservationSpec implements core.Environment.
func (e *Env) ObservationSpec() core.Spec {
	const n = 64
	return core.Spec{Dim: n, Low: make([]float64, n), High: ones(n)}
}

// ones returns a length-n slice of 1s.
func ones(n int) []float64 {
	v := make([]float64, n)
	for i := range v {
		v[i] = 1
	}
	return v
}

var _ core.Environment = (*Env)(nil)
