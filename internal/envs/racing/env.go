package racing

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/danielriddell21/galapagos/internal/core"
)

type Config struct {
	Track   TrackParams
	Car     CarParams
	Sensors SensorParams

	CheckpointReward float64
	SpeedBonus       float64
	TimeBonus        float64
	WallPenalty      float64
}

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

type State struct{ obs []float64 }

func (s State) Observation() []float64 { return s.obs }

type Action struct{ steering, throttle float64 }

func NewAction(steering, throttle float64) Action { return Action{steering, throttle} }

func (a Action) Vector() []float64 { return []float64{a.steering, a.throttle} }

type Env struct {
	cfg   Config
	track *Track
	walls [][2]vec

	bodies []*car
	alive  []bool
	nextCP []int
	passed []int
	best   int
}

func New(cfg Config) *Env { return &Env{cfg: cfg, best: -1} }

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

func (e *Env) hitsWall(prev, pos vec) bool {
	for _, w := range e.walls {
		if segmentsIntersect(prev, pos, w[0], w[1]) {
			return true
		}
	}
	return false
}

func (e *Env) observe(c *car) State {
	rays := sense(c, e.walls, e.cfg.Sensors)
	obs := make([]float64, 0, len(rays)+1)
	obs = append(obs, rays...)
	obs = append(obs, min(c.speed/e.cfg.Car.MaxSpeed, 1))
	return State{obs: obs}
}

func (e *Env) Alive() int {
	n := 0
	for _, a := range e.alive {
		if a {
			n++
		}
	}
	return n
}

func (e *Env) ActionSpec() core.Spec {
	return core.Spec{Dim: 2, Low: []float64{-1, 0}, High: []float64{1, 1}}
}

func (e *Env) ObservationSpec() core.Spec {
	dim := e.cfg.Sensors.Count + 1
	low := make([]float64, dim)
	high := make([]float64, dim)
	for i := range dim {
		high[i] = 1
	}
	return core.Spec{Dim: dim, Low: low, High: high}
}

func (e *Env) Passed(i int) int { return e.passed[i] }

func (e *Env) SetBest(i int) { e.best = i }

func (e *Env) Track() *Track { return e.track }

func (e *Env) Bounds() (minX, minY, maxX, maxY float64) {
	if e.track == nil {
		return 0, 0, 0, 0
	}
	return e.track.Bounds()
}

func (e *Env) CarPosition(i int) core.Vec2 {
	p := e.bodies[i].pos
	return core.Vec2{X: p.X, Y: p.Y}
}

func (e *Env) SensorEndpoints(i int) (origin core.Vec2, ends []core.Vec2) {
	c := e.bodies[i]
	origin = core.Vec2{X: c.pos.X, Y: c.pos.Y}
	for _, p := range rayEndpoints(c, e.walls, e.cfg.Sensors) {
		ends = append(ends, core.Vec2{X: p.X, Y: p.Y})
	}
	return origin, ends
}

func decode(a core.Action) (steering, throttle float64) {
	v := a.Vector()
	if len(v) < 2 {
		return 0, 0
	}
	return math.Tanh(v[0]), (math.Tanh(v[1]) + 1) / 2
}

func angleOf(d vec) float64 { return math.Atan2(d.Y, d.X) }

var _ core.MultiEnvironment = (*Env)(nil)

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
		c := carColor(e.alive[i], i == e.best)
		r.Circle(b.pos.X, b.pos.Y, 6, c)
	}
}

func drawLoop(r core.Renderer, pts []vec, c color.Color) {
	n := len(pts)
	for i := range n {
		j := (i + 1) % n
		r.Line(pts[i].X, pts[i].Y, pts[j].X, pts[j].Y, c)
	}
}

func carColor(alive, best bool) color.Color {
	switch {
	case best:
		return color.RGBA{255, 215, 0, 255} // gold leader
	case !alive:
		return color.RGBA{70, 70, 70, 120} // faded
	default:
		return color.RGBA{80, 180, 255, 220}
	}
}
