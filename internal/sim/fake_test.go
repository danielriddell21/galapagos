package sim

import (
	"iter"
	"math/rand/v2"

	"github.com/danielriddell21/galapagos/internal/core"
)

// fakeState carries a single scalar observation.
type fakeState struct{ x float64 }

func (s fakeState) Observation() []float64 { return []float64{s.x} }

// fakeAction carries a single scalar control.
type fakeAction struct{ v float64 }

func (a fakeAction) Vector() []float64 { return []float64{a.v} }

// fakeIndividual is a member whose constant policy advances a counter; its
// fitness is the number of steps it survives, which is genome[0] steps.
type fakeIndividual struct {
	genome  []float64
	fitness core.Reward
}

func (m *fakeIndividual) Act(s core.State) core.Action { return fakeAction{m.genome[0]} }
func (m *fakeIndividual) Fitness() core.Reward         { return m.fitness }
func (m *fakeIndividual) SetFitness(r core.Reward)     { m.fitness = r }
func (m *fakeIndividual) Genome() []float64            { return m.genome }

// fakePop is a fixed population that never actually evolves; it is enough to
// exercise the loop and evaluation paths.
type fakePop struct {
	members []*fakeIndividual
	gen     int
}

func newFakePop(lifespans ...float64) *fakePop {
	p := &fakePop{}
	for _, l := range lifespans {
		p.members = append(p.members, &fakeIndividual{genome: []float64{l}})
	}
	return p
}

func (p *fakePop) Act(s core.State) core.Action                                             { return p.members[0].Act(s) }
func (p *fakePop) Observe(s core.State, a core.Action, r core.Reward, n core.State, d bool) {}
func (p *fakePop) EndEpisode(total core.Reward)                                             {}
func (p *fakePop) Len() int                                                                 { return len(p.members) }
func (p *fakePop) Generation() int                                                          { return p.gen }
func (p *fakePop) Evolve()                                                                  { p.gen++ }
func (p *fakePop) All() iter.Seq[core.Individual] {
	return func(yield func(core.Individual) bool) {
		for _, m := range p.members {
			if !yield(m) {
				return
			}
		}
	}
}

// fakeMultiEnv earns each body reward 1 per step until it has lived for as many
// steps as the member's action value requests, then marks it done. Because a
// body's outcome depends only on its own member's action, simultaneous and
// per-body rollouts must agree, mirroring the cars-never-interact property.
type fakeMultiEnv struct {
	step  []int
	done  []bool
	alive int
}

func newFakeMultiEnv() *fakeMultiEnv { return &fakeMultiEnv{} }

func (e *fakeMultiEnv) ResetAll(n int, rng *rand.Rand) []core.State {
	e.step = make([]int, n)
	e.done = make([]bool, n)
	e.alive = n
	states := make([]core.State, n)
	for i := range n {
		states[i] = fakeState{0}
	}
	return states
}

func (e *fakeMultiEnv) StepAll(actions []core.Action) ([]core.State, []core.Reward, []bool) {
	states := make([]core.State, len(e.step))
	rewards := make([]core.Reward, len(e.step))
	for i := range e.step {
		if e.done[i] {
			continue
		}
		e.step[i]++
		rewards[i] = 1
		states[i] = fakeState{float64(e.step[i])}
		if float64(e.step[i]) >= actions[i].Vector()[0] {
			e.done[i] = true
			e.alive--
		}
	}
	return states, rewards, e.done
}

func (e *fakeMultiEnv) ActionSpec() core.Spec {
	return core.Spec{Dim: 1, Low: []float64{0}, High: []float64{1}}
}
func (e *fakeMultiEnv) ObservationSpec() core.Spec {
	return core.Spec{Dim: 1, Low: []float64{0}, High: []float64{1}}
}
func (e *fakeMultiEnv) Alive() int             { return e.alive }
func (e *fakeMultiEnv) Render(r core.Renderer) {}
