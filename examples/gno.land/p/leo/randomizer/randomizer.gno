package randomizer

import (
	"time"
)

type Randomizer struct {
	seed int
}

func (r *Randomizer) Next() int {
	const a = 1664525
	const c = 1013904223
	const m = 4294967296

	r.seed = (a*r.seed + c) % m
	return r.seed
}

func NewRandomizer() *Randomizer {
	seed := int(time.Now().Unix())
	return &Randomizer{seed: seed}
}
