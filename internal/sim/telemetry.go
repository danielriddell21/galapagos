package sim

import (
	"log/slog"
	"slices"
)

type Ring[T any] struct {
	buf  []T
	next int
	full bool
}

func NewRing[T any](capacity int) *Ring[T] {
	return &Ring[T]{buf: make([]T, max(capacity, 1))}
}

func (r *Ring[T]) Push(v T) {
	r.buf[r.next] = v
	r.next = (r.next + 1) % len(r.buf)
	if r.next == 0 {
		r.full = true
	}
}

func (r *Ring[T]) Len() int {
	if r.full {
		return len(r.buf)
	}
	return r.next
}

func (r *Ring[T]) Slice() []T {
	if !r.full {
		return slices.Clone(r.buf[:r.next])
	}
	out := make([]T, 0, len(r.buf))
	out = append(out, r.buf[r.next:]...)
	out = append(out, r.buf[:r.next]...)
	return out
}

type GenStats struct {
	Generation int
	Best       float64
	Avg        float64
	Worst      float64
}

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

type Telemetry struct {
	history *Ring[GenStats]
	log     *slog.Logger
}

func NewTelemetry(historyLen int, log *slog.Logger) *Telemetry {
	return &Telemetry{history: NewRing[GenStats](historyLen), log: log}
}

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

func (t *Telemetry) History() []GenStats { return t.history.Slice() }

func (t *Telemetry) BestSeries() []float64 {
	h := t.history.Slice()
	out := make([]float64, len(h))
	for i, s := range h {
		out[i] = s.Best
	}
	return out
}
