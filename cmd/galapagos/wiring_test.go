package main

import (
	"io"
	"log/slog"
	"testing"

	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/spf13/cobra"
)

func seedCmd() *cobra.Command {
	cmd := &cobra.Command{}
	var s int64
	cmd.Flags().Int64Var(&s, "seed", 0, "")
	return cmd
}

func TestResolveSeed(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Explicit --seed wins over the config seed.
	cmd := seedCmd()
	if err := cmd.Flags().Set("seed", "7"); err != nil {
		t.Fatal(err)
	}
	if got := resolveSeed(cmd, 99, log); got != 7 {
		t.Fatalf("explicit flag: got %d, want 7", got)
	}

	// Unset flag falls back to a non-zero config seed.
	if got := resolveSeed(seedCmd(), 99, log); got != 99 {
		t.Fatalf("config seed: got %d, want 99", got)
	}

	// Unset flag and zero config seed yields a random, non-negative seed.
	a := resolveSeed(seedCmd(), 0, log)
	b := resolveSeed(seedCmd(), 0, log)
	if a < 0 || b < 0 {
		t.Fatalf("random seeds must be non-negative, got %d and %d", a, b)
	}
	if a == b {
		t.Fatalf("random seeds should almost never collide, got %d twice", a)
	}
}

func TestDiscreteCount(t *testing.T) {
	tests := []struct {
		name string
		spec core.Spec
		want int
	}{
		{"two actions", core.Spec{Dim: 1, Low: []float64{0}, High: []float64{1}, Discrete: true}, 2},
		{"four actions", core.Spec{Dim: 1, Low: []float64{0}, High: []float64{3}, Discrete: true}, 4},
		{"continuous", core.Spec{Dim: 2, Low: []float64{-1, 0}, High: []float64{1, 1}}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := discreteCount(tt.spec); got != tt.want {
				t.Fatalf("discreteCount = %d, want %d", got, tt.want)
			}
		})
	}
}
