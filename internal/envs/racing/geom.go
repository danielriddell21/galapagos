package racing

import "math"

type vec struct{ X, Y float64 }

func add(a, b vec) vec           { return vec{a.X + b.X, a.Y + b.Y} }
func sub(a, b vec) vec           { return vec{a.X - b.X, a.Y - b.Y} }
func scale(a vec, s float64) vec { return vec{a.X * s, a.Y * s} }
func length(a vec) float64       { return math.Hypot(a.X, a.Y) }

func perp(a vec) vec { return vec{-a.Y, a.X} }

func normalize(a vec) vec {
	l := length(a)
	if l == 0 {
		return vec{}
	}
	return vec{a.X / l, a.Y / l}
}

func cross(a, b, c vec) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}

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

func segmentsIntersect(p1, p2, p3, p4 vec) bool {
	d1 := cross(p3, p4, p1)
	d2 := cross(p3, p4, p2)
	d3 := cross(p1, p2, p3)
	d4 := cross(p1, p2, p4)
	return ((d1 > 0) != (d2 > 0)) && ((d3 > 0) != (d4 > 0))
}

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
