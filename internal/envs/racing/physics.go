package racing

import "math"

// CarParams configures the arcade car model. The dynamics are deliberately
// simple: no tyre model, just heading, speed, friction, and turn rate.
type CarParams struct {
	MaxSpeed float64 // maximum forward speed
	Accel    float64 // acceleration per unit throttle
	TurnRate float64 // maximum heading change per second at full steering
	Drag     float64 // fractional speed lost per second
	Dt       float64 // integration timestep
}

// DefaultCarParams returns a playable arcade configuration.
func DefaultCarParams() CarParams {
	return CarParams{
		MaxSpeed: 320,
		Accel:    260,
		TurnRate: 2.8,
		Drag:     0.6,
		Dt:       1.0 / 60,
	}
}

// car is one vehicle's kinematic state.
type car struct {
	pos     vec
	heading float64 // radians
	speed   float64
}

// dir returns the car's unit forward vector.
func (c *car) dir() vec {
	return vec{math.Cos(c.heading), math.Sin(c.heading)}
}

// advance integrates one timestep given steering in [-1,1] and throttle in
// [0,1], and returns the previous position so the caller can test the swept
// segment for wall collisions.
func (c *car) advance(steering, throttle float64, p CarParams) (prev vec) {
	prev = c.pos
	steering = min(max(steering, -1), 1)
	throttle = min(max(throttle, 0), 1)

	// Turning is more effective with some speed, as in an arcade racer.
	speedFactor := min(c.speed/p.MaxSpeed, 1)
	c.heading += steering * p.TurnRate * p.Dt * (0.3 + 0.7*speedFactor)

	c.speed += throttle * p.Accel * p.Dt
	c.speed -= c.speed * p.Drag * p.Dt
	c.speed = min(max(c.speed, 0), p.MaxSpeed)

	c.pos = add(c.pos, scale(c.dir(), c.speed*p.Dt))
	return prev
}
