package racing

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/danielriddell21/galapagos/internal/core"
)

// Config bundles every parameter of the racing environment. A run is fully
// reproducible from this config plus the RNG passed to ResetAll.
type Config struct {
	Track   TrackParams
	Car     CarParams
	Sensors SensorParams

	CheckpointReward float64 // reward per checkpoint gate crossed
	SpeedBonus       float64 // reward per step scaled by normalized speed
	TimeBonus        float64 // reward per step survived
	WallPenalty      float64 // reward applied once when a car hits a wall
}

// DefaultConfig returns a balanced racing configuration.
func DefaultConfig() Config {
	return Config{
		Track:            DefaultTrackParams(),
		Car:              DefaultCarParams(),
		Sensors:          DefaultSensorParams(),
		CheckpointReward: 10,
		SpeedBonus:       0.01,
		TimeBonus:        0.001,
		WallPenalty:      -5,
	}
}

// State is a car's observation: normalized ray distances followed by normalized
// speed, matching ObservationSpec.
type State struct{ obs []float64 }

// Observation implements core.State.
func (s State) Observation() []float64 { return s.obs }

// Action is a continuous control: {steering ∈ [-1,1], throttle ∈ [0,1]}.
type Action struct{ steering, throttle float64 }

// NewAction builds an action from raw control values.
func NewAction(steering, throttle float64) Action { return Action{steering, throttle} }

// Vector implements core.Action.
func (a Action) Vector() []float64 { return []float64{a.steering, a.throttle} }

// Env is the racing world. It owns one shared track and N car bodies, advancing
// them independently so a whole population can be evaluated at once.
type Env struct {
	cfg   Config
	track *Track
	walls [][2]vec

	bodies []*car
	alive  []bool
	nextCP []int // index of the next checkpoint gate each car must cross
	passed []int // total checkpoints crossed, the primary progress measure
	best   int   // index of the car to highlight when rendering
}

// New returns an empty racing environment; call ResetAll before stepping.
func New(cfg Config) *Env { return &Env{cfg: cfg, best: -1} }

// ResetAll generates the track from rng and places n cars at the start line.
func (e *Env) ResetAll(n int, rng *rand.Rand) []core.State {
	e.track = GenerateTrack(rng, e.cfg.Track)
	e.walls = e.track.walls()
	e.bodies = make([]*car, n)
	e.alive = make([]bool, n)
	e.nextCP = make([]int, n)
	e.passed = make([]int, n)
	e.best = -1
	heading := angleOf(e.track.StartDir)
	states := make([]core.State, n)
	for i := range n {
		e.bodies[i] = &car{pos: e.track.StartPos, heading: heading}
		e.alive[i] = true
		states[i] = e.observe(e.bodies[i])
	}
	return states
}

// StepAll advances every alive car by one timestep.
func (e *Env) StepAll(actions []core.Action) (states []core.State, rewards []core.Reward, dones []bool) {
	n := len(e.bodies)
	states = make([]core.State, n)
	rewards = make([]core.Reward, n)
	dones = make([]bool, n)
	for i := range n {
		if !e.alive[i] {
			dones[i] = true
			continue
		}
		steering, throttle := decode(actions[i])
		body := e.bodies[i]
		prev := body.advance(steering, throttle, e.cfg.Car)

		if e.hitsWall(prev, body.pos) {
			e.alive[i] = false
			rewards[i] = core.Reward(e.cfg.WallPenalty)
			dones[i] = true
			states[i] = e.observe(body)
			continue
		}

		reward := e.cfg.TimeBonus + e.cfg.SpeedBonus*(body.speed/e.cfg.Car.MaxSpeed)
		if e.crossedNextCheckpoint(i, prev, body.pos) {
			reward += e.cfg.CheckpointReward
		}
		rewards[i] = core.Reward(reward)
		states[i] = e.observe(body)
	}
	return states, rewards, dones
}

// crossedNextCheckpoint reports whether car i swept across the gate it was due
// to cross next, and if so advances its progress counter.
func (e *Env) crossedNextCheckpoint(i int, prev, pos vec) bool {
	cps := e.track.Checkpoints
	if len(cps) == 0 {
		return false
	}
	g := cps[e.nextCP[i]]
	if segmentsIntersect(prev, pos, g.A, g.B) {
		e.nextCP[i] = (e.nextCP[i] + 1) % len(cps)
		e.passed[i]++
		return true
	}
	return false
}

// hitsWall reports whether the swept segment prev→pos crosses any wall.
func (e *Env) hitsWall(prev, pos vec) bool {
	for _, w := range e.walls {
		if segmentsIntersect(prev, pos, w[0], w[1]) {
			return true
		}
	}
	return false
}

// observe builds the observation vector for a car.
func (e *Env) observe(c *car) State {
	rays := sense(c, e.walls, e.cfg.Sensors)
	obs := make([]float64, 0, len(rays)+1)
	obs = append(obs, rays...)
	obs = append(obs, min(c.speed/e.cfg.Car.MaxSpeed, 1))
	return State{obs: obs}
}

// Alive implements core.MultiEnvironment.
func (e *Env) Alive() int {
	n := 0
	for _, a := range e.alive {
		if a {
			n++
		}
	}
	return n
}

// ActionSpec implements core.MultiEnvironment.
func (e *Env) ActionSpec() core.Spec {
	return core.Spec{Dim: 2, Low: []float64{-1, 0}, High: []float64{1, 1}}
}

// ObservationSpec implements core.MultiEnvironment.
func (e *Env) ObservationSpec() core.Spec {
	dim := e.cfg.Sensors.Count + 1
	low := make([]float64, dim)
	high := make([]float64, dim)
	for i := range dim {
		high[i] = 1
	}
	return core.Spec{Dim: dim, Low: low, High: high}
}

// Passed returns the number of checkpoints car i has crossed, its primary
// progress measure.
func (e *Env) Passed(i int) int { return e.passed[i] }

// SetBest marks car i as the leader so Render highlights it.
func (e *Env) SetBest(i int) { e.best = i }

// Track exposes the generated track for camera fitting and inspection.
func (e *Env) Track() *Track { return e.track }

// Bounds returns the axis-aligned world bounds of the track for camera fitting.
func (e *Env) Bounds() (minX, minY, maxX, maxY float64) {
	if e.track == nil {
		return 0, 0, 0, 0
	}
	return e.track.Bounds()
}

// CarPosition returns the world position of car i, for camera following.
func (e *Env) CarPosition(i int) core.Vec2 {
	p := e.bodies[i].pos
	return core.Vec2{X: p.X, Y: p.Y}
}

// SensorEndpoints returns the world-space endpoint of each sensor ray for car
// i, for the debug overlay. The shared origin is the car's position.
func (e *Env) SensorEndpoints(i int) (origin core.Vec2, ends []core.Vec2) {
	c := e.bodies[i]
	origin = core.Vec2{X: c.pos.X, Y: c.pos.Y}
	for _, p := range rayEndpoints(c, e.walls, e.cfg.Sensors) {
		ends = append(ends, core.Vec2{X: p.X, Y: p.Y})
	}
	return origin, ends
}

// decode extracts steering and throttle from an action vector, squashing the
// raw control values into their valid ranges: steering via tanh into [-1,1] and
// throttle via a logistic-style map into [0,1].
func decode(a core.Action) (steering, throttle float64) {
	v := a.Vector()
	if len(v) < 2 {
		return 0, 0
	}
	return math.Tanh(v[0]), (math.Tanh(v[1]) + 1) / 2
}

// angleOf returns the heading angle of a direction vector.
func angleOf(d vec) float64 { return math.Atan2(d.Y, d.X) }

var _ core.MultiEnvironment = (*Env)(nil)

// Render draws the track, checkpoints, and cars. Dead cars are faded and the
// leader is highlighted. It only uses core.Renderer primitives.
func (e *Env) Render(r core.Renderer) {
	if e.track == nil {
		return
	}
	wallColor := color.RGBA{120, 120, 130, 255}
	drawLoop(r, e.track.Inner, wallColor)
	drawLoop(r, e.track.Outer, wallColor)
	for _, g := range e.track.Checkpoints {
		r.Line(g.A.X, g.A.Y, g.B.X, g.B.Y, color.RGBA{40, 60, 90, 255})
	}
	for i, b := range e.bodies {
		c := carColor(i, e.alive[i], i == e.best)
		r.Circle(b.pos.X, b.pos.Y, 6, c)
	}
}

// drawLoop draws a closed polyline.
func drawLoop(r core.Renderer, pts []vec, c color.Color) {
	n := len(pts)
	for i := range n {
		j := (i + 1) % n
		r.Line(pts[i].X, pts[i].Y, pts[j].X, pts[j].Y, c)
	}
}

// carColor returns the fill color for a car given its state.
func carColor(i int, alive, best bool) color.Color {
	switch {
	case best:
		return color.RGBA{255, 215, 0, 255} // gold leader
	case !alive:
		return color.RGBA{70, 70, 70, 120} // faded
	default:
		return color.RGBA{80, 180, 255, 220}
	}
}
