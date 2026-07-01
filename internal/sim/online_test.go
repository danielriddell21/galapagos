package sim

import (
	"math/rand/v2"
	"testing"

	"github.com/danielriddell21/galapagos/internal/core"
)

type fakeSingleEnv struct {
	life int
	step int
	init float64
}

func (e *fakeSingleEnv) Reset(rng *rand.Rand) core.State {
	e.step = 0
	e.init = rng.Float64()
	return fakeState{e.init}
}

func (e *fakeSingleEnv) Step(a core.Action) (core.State, core.Reward, bool) {
	e.step++
	return fakeState{float64(e.step)}, 1, e.step >= e.life
}

func (e *fakeSingleEnv) ActionSpec() core.Spec {
	return core.Spec{Dim: 1, Low: []float64{0}, High: []float64{1}}
}

func (e *fakeSingleEnv) ObservationSpec() core.Spec {
	return core.Spec{Dim: 1, Low: []float64{0}, High: []float64{1}}
}
func (e *fakeSingleEnv) Render(r core.Renderer) {}

type countAgent struct{ observed int }

func (a *countAgent) Act(s core.State) core.Action { return fakeAction{1} }
func (a *countAgent) Observe(s core.State, ac core.Action, r core.Reward, n core.State, d bool) {
	a.observed++
}
func (a *countAgent) EndEpisode(total core.Reward) {}

func TestLiveEpisodeMatchesRunEpisode(t *testing.T) {
	batch := RunEpisode(&fakeSingleEnv{life: 5}, &countAgent{}, 100, TrackRNG(42))

	live := NewLiveEpisode(&fakeSingleEnv{life: 5}, &countAgent{}, 100, 42)
	for !live.Step() {
	}
	if float64(batch) != float64(live.Return()) {
		t.Fatalf("live return %v != batch %v", live.Return(), batch)
	}
}

func TestAsMultiSharedTask(t *testing.T) {
	// All bodies of an adapted environment must reset to the same task instance,
	// so their initial observations are identical.
	env := AsMulti(func() core.Environment { return &fakeSingleEnv{life: 10} })
	states := env.ResetAll(5, TrackRNG(42))
	first := states[0].Observation()[0]
	for i, s := range states {
		if s.Observation()[0] != first {
			t.Fatalf("body %d reset to a different task: %v vs %v", i, s.Observation()[0], first)
		}
	}
}
