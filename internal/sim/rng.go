package sim

import "math/rand/v2"

const (
	streamMaster uint64 = 0x0000000000000001
	streamTrack  uint64 = 0x1d8e4e27c47d124f
	streamMember uint64 = 0x2545f4914f6cdd1d
	streamGen    uint64 = 0x3c6ef372fe94f82a
)

func MasterRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewPCG(uint64(seed), streamMaster))
}

func TrackRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewPCG(uint64(seed), streamTrack))
}

func MemberRNG(seed int64, idx int) *rand.Rand {
	return rand.New(rand.NewPCG(uint64(seed)^streamMember, uint64(idx)))
}

func GenRNG(seed int64, gen int) *rand.Rand {
	return rand.New(rand.NewPCG(uint64(seed)^streamGen, uint64(gen)))
}
