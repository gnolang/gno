// Entropy generates fully deterministic, cost-effective, and hard to guess
// numbers.
//
// It is designed both for single-usage, like seeding math/rand or for being
// reused which increases the entropy and its cost effectiveness.
//
// Disclaimer: this package is unsafe and won't prevent others to guess values
// in advance.
//
// It uses the Bernstein's hash djb2 to be CPU-cycle efficient.
package entropy

import (
	"chain/runtime"
	"math"
	"time"
)

type Instance struct {
	value uint32
}

func New() *Instance {
	r := Instance{value: 5381}
	r.addEntropy()
	return &r
}

func FromSeed(seed uint32) *Instance {
	r := Instance{value: seed}
	r.addEntropy()
	return &r
}

func (i *Instance) Seed() uint32 {
	return i.value
}

func (i *Instance) djb2String(input string) {
	for _, c := range input {
		i.djb2Uint32(uint32(c))
	}
}

// super fast random algorithm.
// http://www.cse.yorku.ca/~oz/hash.html
func (i *Instance) djb2Uint32(input uint32) {
	i.value = (i.value << 5) + i.value + input
}

// AddEntropy uses various runtime variables to add entropy to the existing seed.
func (i *Instance) addEntropy() {
	// FIXME: reapply the 5381 initial value?

	// inherit previous entropy
	// nothing to do

	// handle callers
	{
		currentRealm := runtime.CurrentRealm().Address().String()
		i.djb2String(currentRealm)
		originCaller := runtime.OriginCaller().String()
		i.djb2String(originCaller)
	}

	// height
	{
		height := runtime.ChainHeight()
		if height >= math.MaxUint32 {
			height -= math.MaxUint32
		}
		i.djb2Uint32(uint32(height))
	}

	// time
	{
		secs := time.Now().Second()
		i.djb2Uint32(uint32(secs))
		nsecs := time.Now().Nanosecond()
		i.djb2Uint32(uint32(nsecs))
	}

	// FIXME: compute other hard-to-guess but deterministic variables, like real gas?
}

func (i *Instance) Value() uint32 {
	i.addEntropy()
	return i.value
}

func (i *Instance) Value64() uint64 {
	i.addEntropy()
	high := i.value
	i.addEntropy()

	return (uint64(high) << 32) | uint64(i.value)
}
