// Package core defines the interfaces and value types that connect the three
// layers of Galapagos: environments (worlds), agents (learners), and the
// renderer. It imports no other internal package, so environments and agents
// can be written against these contracts without depending on each other or on
// the rendering backend.
package core

import "image/color"

// State is an environment-defined snapshot of the world. Agents only consume
// the flat observation vector returned by [State.Observation].
type State interface {
	// Observation returns the feature vector fed to an agent. All components
	// are normalized to [0,1] by convention so agents are environment-agnostic.
	Observation() []float64
}

// Action is a control signal produced by an agent. Continuous environments use
// the raw vector (for example {steering, throttle}); discrete environments wrap
// an integer choice as a one-element vector.
type Action interface {
	// Vector returns the action as a slice of float64 components.
	Vector() []float64
}

// Reward is the scalar feedback an environment returns for a single step.
type Reward float64

// Spec describes the shape and bounds of an observation or action vector.
type Spec struct {
	Dim      int       // number of components in the vector
	Low      []float64 // inclusive lower bound per component; len == Dim
	High     []float64 // inclusive upper bound per component; len == Dim
	Discrete bool      // true when the vector encodes a discrete choice
}

// Vec2 is a 2-D point or vector in world coordinates.
type Vec2 struct {
	X, Y float64
}

// Renderer is a thin drawing abstraction over the graphics backend. It lives in
// core so environments can implement Render without importing the Ebiten layer.
// All coordinates are in world space; the [Camera] maps them to the screen.
type Renderer interface {
	// Line draws a straight segment between two world points.
	Line(x1, y1, x2, y2 float64, c color.Color)
	// Circle draws a circle of the given world radius centered at (x, y).
	Circle(x, y, radius float64, c color.Color)
	// Polygon draws a closed shape through the given world points.
	Polygon(pts [][2]float64, c color.Color)
	// Text draws a label at the given world position.
	Text(x, y float64, s string)
	// Camera returns the pan/zoom/follow state shared with the backend.
	Camera() *Camera
}
