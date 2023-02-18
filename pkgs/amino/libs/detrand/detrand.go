package detrand

import "math/rand"

// XXX copied from cosmos-sdk simulation.  Make a library?
// DeriveRand derives a new Rand deterministically from another random source.
// Unlike rand.New(rand.NewSource(seed)), the result is "more random"
// depending on the source and state of r.
//
// NOTE: not crypto safe.
func DeriveRand(r *rand.Rand) *rand.Rand {
	const num = 8 // TODO what's a good number?  Too large is too slow.
	ms := multiSource(make([]rand.Source, num))

	for i := 0; i < num; i++ {
		ms[i] = rand.NewSource(r.Int63())
	}

	//nolint:gosec
	return rand.New(ms)
}

type multiSource []rand.Source

func (ms multiSource) Int63() (r int64) {
	for _, source := range ms {
		r ^= source.Int63()
	}

	return r
}

func (ms multiSource) Seed(seed int64) {
	panic("multiSource Seed should not be called")
}
