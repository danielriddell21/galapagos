package core

import "math/rand/v2"

// NewPCG returns a deterministic *rand.Rand backed by a PCG source. The single
// int64 seed is split across the two 64-bit PCG lanes so that the same seed
// always yields the same stream, satisfying the reproducibility requirement.
func NewPCG(seed int64) *rand.Rand {
	lo := uint64(seed)
	hi := lo ^ 0x9e3779b97f4a7c15 // golden-ratio constant decorrelates the lanes
	return rand.New(rand.NewPCG(hi, lo))
}
