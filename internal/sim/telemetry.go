package sim

import (
	"log/slog"
	"slices"
)

// Ring is a fixed-capacity circular buffer that keeps the most recent values.
// It is generic so it can hold fitness samples, frame times, or any series.
type Ring[T any] struct {
	buf  []T
	next int
	full bool
}

// NewRing returns a ring buffer that retains the last capacity values.
func NewRing[T any](capacity int) *Ring[T] {
	return &Ring[T]{buf: make([]T, max(capacity, 1))}
}

// Push appends v, overwriting the oldest value once the buffer is full.
func (r *Ring[T]) Push(v T) {
	r.buf[r.next] = v
	r.next = (r.next + 1) % len(r.buf)
	if r.next == 0 {
		r.full = true
	}
}

// Len returns the number of values currently held.
func (r *Ring[T]) Len() int {
	if r.full {
		return len(r.buf)
	}
	return r.next
}

// Slice returns the held values in insertion order, oldest first.
func (r *Ring[T]) Slice() []T {
	if !r.full {
		return slices.Clone(r.buf[:r.next])
	}
	out := make([]T, 0, len(r.buf))
	out = append(out, r.buf[r.next:]...)
	out = append(out, r.buf[:r.next]...)
	return out
}

// GenStats summarizes one generation's fitness distribution.
type GenStats struct {
	Generation int
	Best       float64
	Avg        float64
	Worst      float64
}

// StatsFrom computes per-generation statistics from a fitness slice. It returns
// the zero value when fitness is empty.
func StatsFrom(generation int, fitness []float64) GenStats {
	if len(fitness) == 0 {
		return GenStats{Generation: generation}
	}
	var sum float64
	for _, f := range fitness {
		sum += f
	}
	return GenStats{
		Generation: generation,
		Best:       slices.Max(fitness),
		Avg:        sum / float64(len(fitness)),
		Worst:      slices.Min(fitness),
	}
}

// Telemetry is the in-memory stats bus the simulation publishes to and the HUD
// reads from. It keeps a bounded history of per-generation statistics for live
// sparklines and logs each generation through slog.
type Telemetry struct {
	history *Ring[GenStats]
	log     *slog.Logger
}

// NewTelemetry returns a telemetry bus retaining the last historyLen
// generations. A nil logger disables logging.
func NewTelemetry(historyLen int, log *slog.Logger) *Telemetry {
	return &Telemetry{history: NewRing[GenStats](historyLen), log: log}
}

// Publish records a generation's statistics and logs a summary line.
func (t *Telemetry) Publish(s GenStats) {
	t.history.Push(s)
	if t.log != nil {
		t.log.Info("generation",
			"gen", s.Generation,
			"best", s.Best,
			"avg", s.Avg,
			"worst", s.Worst,
		)
	}
}

// History returns the retained generation statistics, oldest first.
func (t *Telemetry) History() []GenStats { return t.history.Slice() }

// BestSeries returns the best-fitness value of each retained generation, for
// rendering a fitness-over-time sparkline.
func (t *Telemetry) BestSeries() []float64 {
	h := t.history.Slice()
	out := make([]float64, len(h))
	for i, s := range h {
		out[i] = s.Best
	}
	return out
}
