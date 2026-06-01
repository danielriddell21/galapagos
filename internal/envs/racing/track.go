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
	Points      int     // number of angular control points around the loop
	Radius      float64 // base circuit radius
	Width       float64 // track width
	Displace    float64 // radial amplitude as a fraction of the radius
	SplineSteps int     // samples per control-point span
	Checkpoints int     // number of progress gates
}

// DefaultTrackParams returns reasonable generation parameters.
func DefaultTrackParams() TrackParams {
	return TrackParams{
		Points:      16,
		Radius:      400,
		Width:       90,
		Displace:    0.22,
		SplineSteps: 16,
		Checkpoints: 24,
	}
}

// maxTrackAttempts bounds the deterministic shrink-and-retry loop that
// guarantees a simple (non-self-intersecting) track.
const maxTrackAttempts = 8

// GenerateTrack builds a deterministic, simple closed track from rng and params.
// Control points sit at monotonically increasing angles with smoothed radial
// noise, making the centerline star-shaped; it is smoothed with a Catmull-Rom
// spline and inflated into walls. Because spline overshoot can still fold the
// walls at high amplitude, the angular jitter and radial amplitude are scaled
// down together and the track rebuilt until it no longer self-intersects, with a
// perfect circle as the guaranteed-simple fallback. Cars are therefore never
// boxed in. The same rng state and params always yield the same track.
func GenerateTrack(rng *rand.Rand, p TrackParams) *Track {
	n := max(p.Points, 4)
	spacing := 2 * math.Pi / float64(n)

	jitter := make([]float64, n)
	for i := range n {
		jitter[i] = rng.Float64() - 0.5 // in [-0.5, 0.5]
	}
	raw := make([]float64, n)
	for i := range n {
		raw[i] = rng.Float64()*2 - 1
	}
	noise := make([]float64, n) // smoothed around the ring to avoid spikes
	for i := range n {
		noise[i] = (raw[(i-1+n)%n] + 2*raw[i] + raw[(i+1)%n]) / 4
	}

	build := func(scale float64) *Track {
		cp := make([]vec, n)
		for i := range n {
			ang := spacing*float64(i) + jitter[i]*spacing*0.5*scale
			// Floor the radius at the width so the inner wall stays clear of center.
			r := max(p.Radius*(1+p.Displace*noise[i]*scale), p.Width)
			cp[i] = vec{r * math.Cos(ang), r * math.Sin(ang)}
		}
		t := &Track{Center: catmullClosed(cp, p.SplineSteps), Width: p.Width}
		t.buildWalls()
		t.buildCheckpoints(p.Checkpoints)
		t.setStart()
		return t
	}

	for k := range maxTrackAttempts {
		if t := build(math.Pow(0.7, float64(k))); !t.selfIntersecting() {
			return t
		}
	}
	return build(0) // a circle is always simple
}

// selfIntersecting reports whether the centerline or either wall crosses itself.
func (t *Track) selfIntersecting() bool {
	return selfIntersects(t.Center) || selfIntersects(t.Inner) || selfIntersects(t.Outer)
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
