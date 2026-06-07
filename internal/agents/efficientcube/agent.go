package efficientcube

import (
	"math"

	"github.com/danielriddell21/galapagos/internal/core"
	rubix "github.com/danielriddell21/rubix/pkg/cube"
)

// moveAction carries a cube move as a one-element action vector, matching the
// cube environment's action decoding (it rounds v[0] to a move index).
type moveAction rubix.Move

// Vector implements core.Action.
func (a moveAction) Vector() []float64 { return []float64{float64(a)} }

// Agent solves the cube with beam search over the policy. On the first step of
// an episode it plans a full solution from the cube reconstructed out of the
// observation, then plays it one move per step — so the simulation/GUI animate
// the beam-found solution turning the cube to solved. It implements core.Agent.
type Agent struct {
	p       *Policy
	cfg     BeamConfig
	plan    []rubix.Move
	idx     int
	planned bool
}

// NewAgent returns a beam-search agent with the given search configuration.
func NewAgent(p *Policy, cfg BeamConfig) *Agent {
	return &Agent{p: p, cfg: cfg}
}

// Act plans on first use, then replays the solution move by move.
func (a *Agent) Act(s core.State) core.Action {
	if !a.planned {
		a.plan = a.p.Solve(cubeFromObservation(s.Observation()), a.cfg).Moves
		a.idx = 0
		a.planned = true
	}
	if a.idx >= len(a.plan) {
		return moveAction(0)
	}
	m := a.plan[a.idx]
	a.idx++
	return moveAction(m)
}

// Observe is unused: the policy is trained offline.
func (a *Agent) Observe(s core.State, act core.Action, r core.Reward, next core.State, done bool) {
}

// EndEpisode resets so the next episode re-plans.
func (a *Agent) EndEpisode(total core.Reward) { a.planned = false; a.plan = nil; a.idx = 0 }

// SolutionLen returns the number of moves in the current plan (0 until planned).
func (a *Agent) SolutionLen() int { return len(a.plan) }

// Move returns the index of the move about to be played (for HUD progress).
func (a *Agent) Move() int { return a.idx }

// cubeFromObservation rebuilds the cube from the env's normalized facelet
// observation (each component is color/5, exactly invertible).
func cubeFromObservation(obs []float64) rubix.Cube {
	var f rubix.Facelets
	for i := range min(len(obs), len(f)) {
		f[i] = rubix.Color(int(math.Round(obs[i] * 5)))
	}
	c, err := rubix.FromFacelets(f)
	if err != nil {
		return rubix.Solved()
	}
	return c
}

var _ core.Agent = (*Agent)(nil)
