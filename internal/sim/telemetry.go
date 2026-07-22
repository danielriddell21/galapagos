package sim

import (
	"log/slog"
	"slices"

	"github.com/danielriddell21/crucible/ring"
)

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
	history *ring.Ring[GenStats]
	log     *slog.Logger
}

func NewTelemetry(historyLen int, log *slog.Logger) *Telemetry {
	return &Telemetry{history: ring.New[GenStats](historyLen), log: log}
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
