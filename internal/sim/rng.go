// Package sim owns the simulation loop and wires an environment to an agent and
// a renderer. It is the only layer that holds both an environment and an agent,
// which keeps those two layers decoupled from each other. It also derives all
// random streams, so determinism is defined in one place.
package sim

import "math/rand/v2"

// Distinct stream constants keep independently-seeded RNGs from correlating.
// Each stream is a pure function of the master seed plus, where relevant, an
// index, so results never depend on goroutine scheduling or worker count.
const (
	streamMaster uint64 = 0x0000000000000001
	streamTrack  uint64 = 0x1d8e4e27c47d124f
	streamMember uint64 = 0x2545f4914f6cdd1d
	streamGen    uint64 = 0x3c6ef372fe94f82a
)

// MasterRNG returns the top-level random stream for a run.
func MasterRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewPCG(uint64(seed), streamMaster))
}

// TrackRNG returns the stream used for procedural environment layout, so the
// same seed always produces the same track.
func TrackRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewPCG(uint64(seed), streamTrack))
}

// MemberRNG returns the stream for population member idx. It is derived purely
// from the seed and index, so each member's randomness is independent of the
// order or concurrency with which members are evaluated.
func MemberRNG(seed int64, idx int) *rand.Rand {
	return rand.New(rand.NewPCG(uint64(seed)^streamMember, uint64(idx)))
}

// GenRNG returns the stream used for selection, crossover, and mutation in a
// given generation, so reproduction is deterministic per generation.
func GenRNG(seed int64, gen int) *rand.Rand {
	return rand.New(rand.NewPCG(uint64(seed)^streamGen, uint64(gen)))
}
