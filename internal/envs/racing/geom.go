// Package racing implements the flagship environment: a population of cars on a
// procedurally generated closed-loop track. It implements core.MultiEnvironment
// and depends only on core, so it never imports the renderer backend or any
// agent.
package racing

import (
	"math"
	"slices"
)

// vec is the package's 2-D vector type. It is local (not the shared core type)
// so its literals stay unkeyed without tripping the composites vet check.
type vec struct{ X, Y float64 }

func add(a, b vec) vec           { return vec{a.X + b.X, a.Y + b.Y} }
func sub(a, b vec) vec           { return vec{a.X - b.X, a.Y - b.Y} }
func scale(a vec, s float64) vec { return vec{a.X * s, a.Y * s} }
func dot(a, b vec) float64       { return a.X*b.X + a.Y*b.Y }
func length(a vec) float64       { return math.Hypot(a.X, a.Y) }

// perp returns the left-hand perpendicular of a.
func perp(a vec) vec { return vec{-a.Y, a.X} }

// normalize returns a unit vector in the direction of a, or the zero vector if
// a has zero length.
func normalize(a vec) vec {
	l := length(a)
	if l == 0 {
		return vec{}
	}
	return vec{a.X / l, a.Y / l}
}

// cross returns the z-component of the 2-D cross product of (b-a) and (c-a),
// which is positive when a→b→c turns left.
func cross(a, b, c vec) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}

// convexHull returns the convex hull of pts in counter-clockwise order using
// Andrew's monotone chain. The result has no repeated endpoint.
func convexHull(pts []vec) []vec {
	if len(pts) < 3 {
		return slices.Clone(pts)
	}
	p := slices.Clone(pts)
	slices.SortFunc(p, func(a, b vec) int {
		if a.X != b.X {
			return cmp(a.X, b.X)
		}
		return cmp(a.Y, b.Y)
	})
	build := func(points []vec) []vec {
		var h []vec
		for _, pt := range points {
			for len(h) >= 2 && cross(h[len(h)-2], h[len(h)-1], pt) <= 0 {
				h = h[:len(h)-1]
			}
			h = append(h, pt)
		}
		return h[:len(h)-1] // drop last point; it repeats the next chain's start
	}
	lower := build(p)
	slices.Reverse(p)
	upper := build(p)
	return append(lower, upper...)
}

func cmp(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// rayHit returns the distance along a ray (from origin in unit direction dir)
// to its intersection with segment a→b, and whether they intersect ahead of the
// origin. Distances behind the origin are ignored.
func rayHit(origin, dir, a, b vec) (dist float64, ok bool) {
	e := sub(b, a)
	denom := dir.X*e.Y - dir.Y*e.X
	if math.Abs(denom) < 1e-12 {
		return 0, false // parallel
	}
	diff := sub(a, origin)
	t := (diff.X*e.Y - diff.Y*e.X) / denom     // distance along the ray
	u := (diff.X*dir.Y - diff.Y*dir.X) / denom // position along the segment
	if t >= 0 && u >= 0 && u <= 1 {
		return t, true
	}
	return 0, false
}

// segmentsIntersect reports whether segments p1→p2 and p3→p4 cross.
func segmentsIntersect(p1, p2, p3, p4 vec) bool {
	d1 := cross(p3, p4, p1)
	d2 := cross(p3, p4, p2)
	d3 := cross(p1, p2, p3)
	d4 := cross(p1, p2, p4)
	return ((d1 > 0) != (d2 > 0)) && ((d3 > 0) != (d4 > 0))
}
