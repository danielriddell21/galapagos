package sim

import (
	"math/rand/v2"

	"github.com/danielriddell21/galapagos/internal/core"
)

// multiAdapter presents a collection of independent single-agent environments
// as one MultiEnvironment, so population agents can train on environments such
// as cart-pole and maze without any change to the agent or the loop.
type multiAdapter struct {
	make   func() core.Environment
	bodies []core.Environment
	done   []bool
	alive  int
}

// AsMulti adapts a single-agent environment constructor into a MultiEnvironment.
// Each body is an independent environment instance, mirroring the per-member
// rollout used for the racing population so the determinism guarantees hold.
func AsMulti(make func() core.Environment) core.MultiEnvironment {
	return &multiAdapter{make: make}
}

func (m *multiAdapter) ResetAll(n int, rng *rand.Rand) []core.State {
	m.bodies = make([]core.Environment, n)
	m.done = make([]bool, n)
	m.alive = n
	// All bodies share one task instance, like the racing population shares one
	// track: every body resets from the same derived seed, so members are scored
	// on identical conditions and the swarm is coherent to watch.
	task := rng.Uint64()
	states := make([]core.State, n)
	for i := range n {
		m.bodies[i] = m.make()
		states[i] = m.bodies[i].Reset(rand.New(rand.NewPCG(task, 0)))
	}
	return states
}

func (m *multiAdapter) StepAll(actions []core.Action) (states []core.State, rewards []core.Reward, dones []bool) {
	states = make([]core.State, len(m.bodies))
	rewards = make([]core.Reward, len(m.bodies))
	for i, env := range m.bodies {
		if m.done[i] {
			states[i] = nil
			continue
		}
		s, r, done := env.Step(actions[i])
		states[i], rewards[i] = s, r
		if done {
			m.done[i] = true
			m.alive--
		}
	}
	return states, rewards, m.done
}

func (m *multiAdapter) ActionSpec() core.Spec      { return m.make().ActionSpec() }
func (m *multiAdapter) ObservationSpec() core.Spec { return m.make().ObservationSpec() }
func (m *multiAdapter) Alive() int                 { return m.alive }

func (m *multiAdapter) Render(r core.Renderer) {
	for i, env := range m.bodies {
		if !m.done[i] {
			env.Render(r)
		}
	}
}

// Bounds forwards the world bounds of a representative body, so a population
// running on an adapted single-agent environment can still fit the camera. It
// reports false when the underlying environment does not expose bounds.
func (m *multiAdapter) Bounds() (minX, minY, maxX, maxY float64, ok bool) {
	var probe core.Environment
	if len(m.bodies) > 0 {
		probe = m.bodies[0]
	} else {
		probe = m.make()
	}
	if b, has := probe.(interface {
		Bounds() (float64, float64, float64, float64)
	}); has {
		minX, minY, maxX, maxY = b.Bounds()
		return minX, minY, maxX, maxY, true
	}
	return 0, 0, 0, 0, false
}

var _ core.MultiEnvironment = (*multiAdapter)(nil)
