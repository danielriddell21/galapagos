// Package racing implements the flagship environment: a population of cars on a
// procedurally generated closed-loop track. It implements core.MultiEnvironment
// and depends only on core, so it never imports the renderer backend or any
// agent.
package racing

import "math"

// vec is the package's 2-D vector type. It is local (not the shared core type)
// so its literals stay unkeyed without tripping the composites vet check.
type vec struct{ X, Y float64 }

func add(a, b vec) vec           { return vec{a.X + b.X, a.Y + b.Y} }
func sub(a, b vec) vec           { return vec{a.X - b.X, a.Y - b.Y} }
func scale(a vec, s float64) vec { return vec{a.X * s, a.Y * s} }
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

// selfIntersects reports whether the closed polyline poly crosses itself. Pairs
// of segments that share a vertex (adjacent, including across the wrap) are
// skipped, since they meet by construction rather than truly intersecting.
func selfIntersects(poly []vec) bool {
	n := len(poly)
	for i := range n {
		a1, a2 := poly[i], poly[(i+1)%n]
		for j := i + 2; j < n; j++ {
			if i == 0 && j == n-1 {
				continue // first and last segments are adjacent across the wrap
			}
			b1, b2 := poly[j], poly[(j+1)%n]
			if segmentsIntersect(a1, a2, b1, b2) {
				return true
			}
		}
	}
	return false
}
