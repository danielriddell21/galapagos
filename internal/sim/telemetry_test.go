package sim

import (
	"slices"
	"testing"
)

func TestStatsFrom(t *testing.T) {
	got := StatsFrom(2, []float64{1, 2, 3, 6})
	want := GenStats{Generation: 2, Best: 6, Avg: 3, Worst: 1}
	if got != want {
		t.Fatalf("StatsFrom = %+v, want %+v", got, want)
	}
}

func TestTelemetryBestSeries(t *testing.T) {
	tel := NewTelemetry(8, nil)
	tel.Publish(StatsFrom(0, []float64{1, 2}))
	tel.Publish(StatsFrom(1, []float64{3, 5}))
	if got, want := tel.BestSeries(), []float64{2, 5}; !slices.Equal(got, want) {
		t.Fatalf("BestSeries = %v, want %v", got, want)
	}
}
