package racing

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/danielriddell21/galapagos/internal/core"
)

func newRNG() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

func TestRayHit(t *testing.T) {
	// Ray from origin along +x hits the vertical segment x=5 at distance 5.
	d, ok := rayHit(vec{0, 0}, vec{1, 0}, vec{5, -1}, vec{5, 1})
	if !ok || math.Abs(d-5) > 1e-9 {
		t.Fatalf("rayHit = (%g,%v), want (5,true)", d, ok)
	}
	// A ray pointing away from the segment must not hit.
	if _, ok := rayHit(vec{0, 0}, vec{-1, 0}, vec{5, -1}, vec{5, 1}); ok {
		t.Fatal("ray pointing away should not hit")
	}
}

func TestSegmentsIntersect(t *testing.T) {
	tests := []struct {
		name           string
		p1, p2, p3, p4 vec
		want           bool
	}{
		{"cross", vec{0, 0}, vec{2, 2}, vec{0, 2}, vec{2, 0}, true},
		{"parallel", vec{0, 0}, vec{2, 0}, vec{0, 1}, vec{2, 1}, false},
		{"disjoint", vec{0, 0}, vec{1, 0}, vec{2, 0}, vec{3, 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := segmentsIntersect(tt.p1, tt.p2, tt.p3, tt.p4); got != tt.want {
				t.Fatalf("segmentsIntersect = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvexHullSquare(t *testing.T) {
	pts := []vec{{0, 0}, {4, 0}, {4, 4}, {0, 4}, {2, 2}} // interior point ignored
	hull := convexHull(pts)
	if len(hull) != 4 {
		t.Fatalf("hull has %d points, want 4", len(hull))
	}
}

func TestGenerateTrackDeterministic(t *testing.T) {
	p := DefaultTrackParams()
	a := GenerateTrack(newRNG(), p)
	b := GenerateTrack(newRNG(), p)
	if len(a.Center) != len(b.Center) {
		t.Fatalf("center lengths differ: %d vs %d", len(a.Center), len(b.Center))
	}
	for i := range a.Center {
		if a.Center[i] != b.Center[i] {
			t.Fatalf("centerline differs at %d: %v vs %v", i, a.Center[i], b.Center[i])
		}
	}
}

func TestTrackWallSpacing(t *testing.T) {
	tr := GenerateTrack(newRNG(), DefaultTrackParams())
	if len(tr.Inner) != len(tr.Center) || len(tr.Outer) != len(tr.Center) {
		t.Fatalf("wall lengths must match centerline")
	}
	// Each centerline point sits roughly halfway between its walls.
	for i := range tr.Center {
		d := length(sub(tr.Outer[i], tr.Inner[i]))
		if math.Abs(d-tr.Width) > 1e-6 {
			t.Fatalf("wall span at %d = %g, want %g", i, d, tr.Width)
		}
	}
}

func TestObservationWithinBounds(t *testing.T) {
	e := New(DefaultConfig())
	states := e.ResetAll(4, newRNG())
	spec := e.ObservationSpec()
	for _, s := range states {
		obs := s.Observation()
		if len(obs) != spec.Dim {
			t.Fatalf("obs len = %d, want %d", len(obs), spec.Dim)
		}
		for j, v := range obs {
			if v < spec.Low[j] || v > spec.High[j] {
				t.Fatalf("obs[%d] = %g out of [%g,%g]", j, v, spec.Low[j], spec.High[j])
			}
		}
	}
}

func TestStepRewardsAndAliveCount(t *testing.T) {
	e := New(DefaultConfig())
	e.ResetAll(3, newRNG())
	if e.Alive() != 3 {
		t.Fatalf("alive after reset = %d, want 3", e.Alive())
	}
	actions := []core.Action{NewAction(0, 1), NewAction(0, 1), NewAction(0, 1)}
	_, rewards, _ := e.StepAll(actions)
	if len(rewards) != 3 {
		t.Fatalf("rewards len = %d, want 3", len(rewards))
	}
}

func TestCarCrashesIntoWall(t *testing.T) {
	// Hard-turning at full throttle drives a car into a wall within the step
	// budget, proving collisions terminate a body.
	e := New(DefaultConfig())
	e.ResetAll(1, newRNG())
	crashed := false
	for range 2000 {
		_, _, dones := e.StepAll([]core.Action{NewAction(1, 1)})
		if dones[0] {
			crashed = true
			break
		}
	}
	if !crashed {
		t.Fatal("a full-lock full-throttle car should eventually hit a wall")
	}
	if e.Alive() != 0 {
		t.Fatalf("alive = %d after crash, want 0", e.Alive())
	}
}

func TestSensorAnglesSpan(t *testing.T) {
	s := SensorParams{Count: 7, FOV: math.Pi, MaxRange: 100}
	angles := s.rayAngles()
	if got := slices.Min(angles); math.Abs(got+math.Pi/2) > 1e-9 {
		t.Fatalf("min angle = %g, want -pi/2", got)
	}
	if got := slices.Max(angles); math.Abs(got-math.Pi/2) > 1e-9 {
		t.Fatalf("max angle = %g, want pi/2", got)
	}
}
