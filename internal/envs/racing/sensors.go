package racing

import "math"

// SensorParams configures the raycast sensors fanned from the car's nose.
type SensorParams struct {
	Count    int     // number of rays
	FOV      float64 // total fan angle in radians, centered on the heading
	MaxRange float64 // distance at which a ray reads 1 (no wall seen)
}

// DefaultSensorParams returns a 7-ray, 180-degree fan.
func DefaultSensorParams() SensorParams {
	return SensorParams{Count: 7, FOV: math.Pi, MaxRange: 600}
}

// rayAngles returns the per-ray offsets from the car heading, spread evenly
// across the field of view.
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

// rayEndpoints returns the world-space endpoint of each sensor ray, clamped to
// the maximum range, for the debug overlay.
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

// sense casts the sensor rays from the car and returns each ray's normalized
// distance to the nearest wall in [0,1], where 1 means nothing within range.
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
