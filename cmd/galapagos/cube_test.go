package main

import "testing"

func TestCoupleScramble(t *testing.T) {
	tests := []struct {
		name             string
		target           int
		scrK, maxDepth   int
		wantK, wantDepth int
	}{
		{"raises shallow training to the target", 8, 6, 18, 8, 18},
		{"raises a tight horizon above the target", 10, 6, 8, 10, 12},
		{"keeps deeper user-set values", 6, 12, 30, 12, 30},
		{"leaves matching values untouched", 8, 8, 10, 8, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, depth := coupleScramble(tt.target, tt.scrK, tt.maxDepth)
			if k != tt.wantK || depth != tt.wantDepth {
				t.Fatalf("coupleScramble(%d, %d, %d) = %d, %d; want %d, %d",
					tt.target, tt.scrK, tt.maxDepth, k, depth, tt.wantK, tt.wantDepth)
			}
		})
	}
}
