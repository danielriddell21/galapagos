package racing

import (
	"math"
	"math/rand/v2"
)

// Track is a procedurally generated closed-loop circuit. All polylines are
// ordered counter-clockwise; the centerline, inner, and outer walls share the
// same length and index, so center[i] lies between inner[i] and outer[i].
type Track struct {
	Center      []vec   // smoothed centerline, one loop, no repeated endpoint
	Inner       []vec   // inner wall, offset from the centerline
	Outer       []vec   // outer wall, offset from the centerline
	Checkpoints []gate  // progress gates spanning inner to outer
	Width       float64 // total track width
	StartPos    vec     // start/finish position on the centerline
	StartDir    vec     // unit heading at the start, along the centerline
}

// gate is a checkpoint: a segment from the inner to the outer wall that a car
// crosses to register progress around the loop.
type gate struct {
	A, B vec
}

// TrackParams controls procedural track generation.
type TrackParams struct {
	Points      int     // number of random seed points before hulling
	Radius      float64 // approximate circuit radius
	Width       float64 // track width
	Displace    float64 // midpoint displacement magnitude
	SplineSteps int     // samples per control-point span
	Checkpoints int     // number of progress gates
}

// DefaultTrackParams returns reasonable generation parameters.
func DefaultTrackParams() TrackParams {
	return TrackParams{
		Points:      12,
		Radius:      400,
		Width:       90,
		Displace:    0.35,
		SplineSteps: 16,
		Checkpoints: 24,
	}
}

// GenerateTrack builds a deterministic track from rng and params: random points
// → convex hull → midpoint displacement → Catmull-Rom smoothing → wall offset →
// checkpoints. The same rng state and params always yield the same track.
func GenerateTrack(rng *rand.Rand, p TrackParams) *Track {
	// Scatter points in a disc, biased outward so the hull is a fair loop.
	pts := make([]vec, p.Points)
	for i := range p.Points {
		ang := rng.Float64() * 2 * math.Pi
		r := p.Radius * (0.6 + 0.4*rng.Float64())
		pts[i] = vec{r * math.Cos(ang), r * math.Sin(ang)}
	}

	hull := convexHull(pts)
	control := displaceMidpoints(hull, rng, p.Displace)
	center := catmullClosed(control, p.SplineSteps)

	t := &Track{Center: center, Width: p.Width}
	t.buildWalls()
	t.buildCheckpoints(p.Checkpoints)
	t.setStart()
	return t
}

// displaceMidpoints inserts a perturbed midpoint between each pair of adjacent
// control points, pushing it along the edge normal to create organic curves.
func displaceMidpoints(loop []vec, rng *rand.Rand, mag float64) []vec {
	n := len(loop)
	out := make([]vec, 0, n*2)
	for i := range n {
		a := loop[i]
		b := loop[(i+1)%n]
		out = append(out, a)
		mid := scale(add(a, b), 0.5)
		edge := sub(b, a)
		normal := normalize(perp(edge))
		offset := (rng.Float64()*2 - 1) * mag * length(edge)
		out = append(out, add(mid, scale(normal, offset)))
	}
	return out
}

// catmullClosed samples a closed Catmull-Rom spline through the control points.
func catmullClosed(cp []vec, steps int) []vec {
	n := len(cp)
	if n < 4 || steps < 1 {
		return cp
	}
	out := make([]vec, 0, n*steps)
	for i := range n {
		p0 := cp[(i-1+n)%n]
		p1 := cp[i]
		p2 := cp[(i+1)%n]
		p3 := cp[(i+2)%n]
		for s := range steps {
			out = append(out, catmullRom(p0, p1, p2, p3, float64(s)/float64(steps)))
		}
	}
	return out
}

// catmullRom evaluates the centripetal-style uniform Catmull-Rom spline at t in
// [0,1] for the span p1→p2.
func catmullRom(p0, p1, p2, p3 vec, t float64) vec {
	t2 := t * t
	t3 := t2 * t
	a := scale(p0, -0.5*t3+t2-0.5*t)
	b := scale(p1, 1.5*t3-2.5*t2+1)
	c := scale(p2, -1.5*t3+2*t2+0.5*t)
	d := scale(p3, 0.5*t3-0.5*t2)
	return add(add(a, b), add(c, d))
}

// buildWalls offsets the centerline by half the width along its local normal.
func (t *Track) buildWalls() {
	n := len(t.Center)
	t.Inner = make([]vec, n)
	t.Outer = make([]vec, n)
	half := t.Width / 2
	for i := range n {
		prev := t.Center[(i-1+n)%n]
		next := t.Center[(i+1)%n]
		normal := normalize(perp(sub(next, prev)))
		t.Inner[i] = sub(t.Center[i], scale(normal, half))
		t.Outer[i] = add(t.Center[i], scale(normal, half))
	}
}

// buildCheckpoints places count evenly spaced gates around the loop.
func (t *Track) buildCheckpoints(count int) {
	n := len(t.Center)
	if count <= 0 || n == 0 {
		return
	}
	t.Checkpoints = make([]gate, 0, count)
	for k := range count {
		i := k * n / count
		t.Checkpoints = append(t.Checkpoints, gate{A: t.Inner[i], B: t.Outer[i]})
	}
}

// setStart places the start/finish at the first centerline point.
func (t *Track) setStart() {
	if len(t.Center) < 2 {
		return
	}
	t.StartPos = t.Center[0]
	t.StartDir = normalize(sub(t.Center[1], t.Center[0]))
}

// Bounds returns the axis-aligned bounding box of the track's outer wall, for
// fitting the camera to the whole circuit.
func (t *Track) Bounds() (minX, minY, maxX, maxY float64) {
	if len(t.Outer) == 0 {
		return 0, 0, 0, 0
	}
	minX, minY = t.Outer[0].X, t.Outer[0].Y
	maxX, maxY = minX, minY
	for _, p := range t.Outer {
		minX, maxX = min(minX, p.X), max(maxX, p.X)
		minY, maxY = min(minY, p.Y), max(maxY, p.Y)
	}
	return minX, minY, maxX, maxY
}

// walls returns the inner and outer wall segments for collision and raycasting.
func (t *Track) walls() [][2]vec {
	n := len(t.Center)
	segs := make([][2]vec, 0, n*2)
	for i := range n {
		j := (i + 1) % n
		segs = append(segs, [2]vec{t.Inner[i], t.Inner[j]})
		segs = append(segs, [2]vec{t.Outer[i], t.Outer[j]})
	}
	return segs
}
