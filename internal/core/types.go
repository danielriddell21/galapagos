package core

import "image/color"

type State interface {
	Observation() []float64
}

type Action interface {
	Vector() []float64
}

type Reward float64

type Spec struct {
	Dim      int
	Low      []float64
	High     []float64
	Discrete bool
}

type Vec2 struct {
	X, Y float64
}

type Renderer interface {
	Line(x1, y1, x2, y2 float64, c color.Color)

	Circle(x, y, radius float64, c color.Color)

	Polygon(pts [][2]float64, c color.Color)

	Text(x, y float64, s string)

	Camera() *Camera
}
