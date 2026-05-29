package core

import (
	"math"
	"testing"
)

func TestCameraRoundTrip(t *testing.T) {
	c := NewCamera(800, 600)
	c.Follow(10, -4)
	c.Zoom = 2.5
	for _, p := range []Vec2{{0, 0}, {100, 50}, {-30, 12}} {
		sx, sy := c.WorldToScreen(p.X, p.Y)
		wx, wy := c.ScreenToWorld(sx, sy)
		if math.Abs(wx-p.X) > 1e-9 || math.Abs(wy-p.Y) > 1e-9 {
			t.Fatalf("round trip failed for %v: got (%g,%g)", p, wx, wy)
		}
	}
}

func TestCameraFitBounds(t *testing.T) {
	c := NewCamera(800, 600)
	c.FitBounds(-100, -100, 100, 100, 0)
	if c.X != 0 || c.Y != 0 {
		t.Fatalf("center = (%g,%g), want origin", c.X, c.Y)
	}
	// The 200-wide world must fit the 600-tall, 800-wide viewport via height.
	if want := 600.0 / 200.0; math.Abs(c.Zoom-want) > 1e-9 {
		t.Fatalf("zoom = %g, want %g", c.Zoom, want)
	}
}
