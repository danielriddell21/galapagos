package core

import "math/rand/v2"

func NewPCG(seed int64) *rand.Rand {
	lo := uint64(seed)
	hi := lo ^ 0x9e3779b97f4a7c15 // golden-ratio constant decorrelates the lanes
	return rand.New(rand.NewPCG(hi, lo))
}
