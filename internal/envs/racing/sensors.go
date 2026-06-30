package racing

import "math"

type SensorParams struct {
	Count    int
	FOV      float64
	MaxRange float64
}

func DefaultSensorParams() SensorParams {
	return SensorParams{Count: 7, FOV: math.Pi, MaxRange: 600}
}

func (s SensorParams) rayAngles() []float64 {
	angles := make([]float64, s.Count)
	if s.Count == 1 {
		return angles
	}
	step := s.FOV / float64(s.Count-1)
	for i := range s.Count {
		angles[i] = -s.FOV/2 + step*float64(i)
	}
	return angles
}

func rayEndpoints(c *car, walls [][2]vec, s SensorParams) []vec {
	out := make([]vec, s.Count)
	for i, off := range s.rayAngles() {
		ang := c.heading + off
		dir := vec{math.Cos(ang), math.Sin(ang)}
		nearest := s.MaxRange
		for _, w := range walls {
			if d, ok := rayHit(c.pos, dir, w[0], w[1]); ok && d < nearest {
				nearest = d
			}
		}
		out[i] = add(c.pos, scale(dir, nearest))
	}
	return out
}

func sense(c *car, walls [][2]vec, s SensorParams) []float64 {
	out := make([]float64, s.Count)
	for i, off := range s.rayAngles() {
		ang := c.heading + off
		dir := vec{math.Cos(ang), math.Sin(ang)}
		nearest := s.MaxRange
		for _, w := range walls {
			if d, ok := rayHit(c.pos, dir, w[0], w[1]); ok && d < nearest {
				nearest = d
			}
		}
		out[i] = min(nearest/s.MaxRange, 1)
	}
	return out
}
