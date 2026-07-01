package racing

import (
	"math"
	"math/rand/v2"
)

type Track struct {
	Center      []vec
	Inner       []vec
	Outer       []vec
	Checkpoints []gate
	Width       float64
	StartPos    vec
	StartDir    vec
}

type gate struct {
	A, B vec
}

type TrackParams struct {
	Points      int
	Radius      float64
	Width       float64
	Displace    float64
	SplineSteps int
	Checkpoints int
}

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

const maxTrackAttempts = 8

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

func (t *Track) selfIntersecting() bool {
	return selfIntersects(t.Center) || selfIntersects(t.Inner) || selfIntersects(t.Outer)
}

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

func catmullRom(p0, p1, p2, p3 vec, t float64) vec {
	t2 := t * t
	t3 := t2 * t
	a := scale(p0, -0.5*t3+t2-0.5*t)
	b := scale(p1, 1.5*t3-2.5*t2+1)
	c := scale(p2, -1.5*t3+2*t2+0.5*t)
	d := scale(p3, 0.5*t3-0.5*t2)
	return add(add(a, b), add(c, d))
}

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

func (t *Track) setStart() {
	if len(t.Center) < 2 {
		return
	}
	t.StartPos = t.Center[0]
	t.StartDir = normalize(sub(t.Center[1], t.Center[0]))
}

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
