package racing

import "math"

type CarParams struct {
	MaxSpeed float64
	Accel    float64
	TurnRate float64
	Drag     float64
	Dt       float64
}

func DefaultCarParams() CarParams {
	return CarParams{
		MaxSpeed: 320,
		Accel:    260,
		TurnRate: 2.8,
		Drag:     0.6,
		Dt:       1.0 / 60,
	}
}

type car struct {
	pos     vec
	heading float64
	speed   float64
}

func (c *car) dir() vec {
	return vec{math.Cos(c.heading), math.Sin(c.heading)}
}

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
