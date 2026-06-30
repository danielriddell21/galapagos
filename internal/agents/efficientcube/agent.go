package efficientcube

import (
	"math"

	rubix "github.com/danielriddell21/rubix/pkg/cube"

	"github.com/danielriddell21/galapagos/internal/core"
)

type moveAction rubix.Move

func (a moveAction) Vector() []float64 { return []float64{float64(a)} }

type Agent struct {
	p       *Policy
	cfg     BeamConfig
	plan    []rubix.Move
	idx     int
	planned bool
}

func NewAgent(p *Policy, cfg BeamConfig) *Agent {
	return &Agent{p: p, cfg: cfg}
}

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

func (a *Agent) Observe(s core.State, act core.Action, r core.Reward, next core.State, done bool) {
}

func (a *Agent) EndEpisode(total core.Reward) { a.planned = false; a.plan = nil; a.idx = 0 }

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
