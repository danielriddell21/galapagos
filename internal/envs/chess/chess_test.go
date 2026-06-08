package chess

import (
	"math/rand/v2"
	"testing"

	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/render"
	gambit "github.com/danielriddell21/gambit/pkg/chess"
)

func newRNG() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

// prefs is an action vector of move preferences.
type prefs []float64

func (p prefs) Vector() []float64 { return p }

// toSquare wants the agent to move a piece to the given square: it sets that
// "to" preference high and leaves the rest at zero.
func toSquare(s gambit.Square) prefs {
	v := make(prefs, 128)
	v[64+int(s)] = 1
	return v
}

func TestResetStartsStandardPosition(t *testing.T) {
	s := New(DefaultConfig()).Reset(newRNG()).(State)
	if s.b.FEN() != gambit.StartingFEN {
		t.Fatalf("reset FEN = %q, want the starting position", s.b.FEN())
	}
}

func TestObservationWithinBounds(t *testing.T) {
	s := New(DefaultConfig()).Reset(newRNG())
	obs := s.Observation()
	if len(obs) != 64 {
		t.Fatalf("observation len = %d, want 64", len(obs))
	}
	for _, v := range obs {
		if v < 0 || v > 1 {
			t.Fatalf("observation %g out of [0,1]", v)
		}
	}
}

func TestActionSpecIsMovePreferences(t *testing.T) {
	spec := New(DefaultConfig()).ActionSpec()
	if spec.Dim != 128 {
		t.Fatalf("action dim = %d, want 128", spec.Dim)
	}
}

func TestBestLegalMoveHonorsPreference(t *testing.T) {
	// From the opening, steer the agent to play a pawn to e4.
	b := gambit.NewStartingBoard()
	e4, _ := gambit.ParseSquare("e4")
	m, ok := bestLegalMove(b, toSquare(e4))
	if !ok {
		t.Fatal("no legal move found from the opening")
	}
	if m.To() != e4 {
		t.Fatalf("preferred move went to %v, want e4", m.To())
	}
}

func TestBestLegalMoveTerminalHasNone(t *testing.T) {
	// Fool's mate: White is checkmated, so White has no legal move.
	b, err := gambit.ParseFEN("rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := bestLegalMove(b, prefs(make([]float64, 128))); ok {
		t.Fatal("expected no legal move in a checkmated position")
	}
}

func TestStepCaptureRewardsMaterial(t *testing.T) {
	// White pawn on e4 can capture a black pawn on d5; steering there must yield
	// a positive shaped reward.
	e := New(Config{MaxPlies: 20})
	e.Reset(newRNG())
	b, err := gambit.ParseFEN("rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 2")
	if err != nil {
		t.Fatal(err)
	}
	e.b = b
	d5, _ := gambit.ParseSquare("d5")
	_, r, _ := e.Step(toSquare(d5))
	if r <= 0 {
		t.Fatalf("capturing a pawn gave reward %v, want > 0", float64(r))
	}
}

func TestStepCheckmateWins(t *testing.T) {
	// Back-rank mate: White rook on a1 plays Ra8#; the black king on g8 is boxed
	// in by its own pawns on f7, g7, h7.
	e := New(Config{MaxPlies: 20})
	e.Reset(newRNG())
	b, err := gambit.ParseFEN("6k1/5ppp/8/8/8/8/8/R6K w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	e.b = b
	a8, _ := gambit.ParseSquare("a8")
	_, r, done := e.Step(toSquare(a8))
	if !done || r != 1 {
		t.Fatalf("checkmating move = (reward %v, done %v), want (1, true)", float64(r), done)
	}
}

func TestEpisodeEndsAtPlyBudget(t *testing.T) {
	e := New(Config{MaxPlies: 4}) // random vs random, bounded
	e.Reset(newRNG())
	var done bool
	for range 4 {
		// A near-uniform preference lets the agent pick some legal move.
		_, _, done = e.Step(prefs(make([]float64, 128)))
		if done {
			break
		}
	}
	if !done {
		t.Fatal("episode should end at the ply budget")
	}
}

func TestRenderDoesNotPanic(t *testing.T) {
	e := New(DefaultConfig())
	e.Reset(newRNG())
	e.Step(prefs(make([]float64, 128)))
	e.Render(render.NewNop()) // exercises the board + piece draw path
}

var _ core.Environment = (*Env)(nil)
