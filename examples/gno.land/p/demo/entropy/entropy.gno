// Entropy generates fully deterministic, cost-effective, and hard to guess
// numbers.
//
// It is designed both for for single-usage, like seeding math/rand or for being
// reused which increases the entropy and its cost effectiveness.
//
// Disclaimer: this package is unsafe and won't prevent others to guess values
// in advance.
//
// It uses the Bernstein's hash djb2 to be CPU-cycle efficient.
package entropy

import (
	"std"
	"time"
)

type Instance struct {
	value int64
}

func New() *Instance {
	r := Instance{value: 5381}
	r.addEntropy()
	return &r
}

func FromSeed(seed int64) *Instance {
	r := Instance{value: seed}
	r.addEntropy()
	return &r
}

func (i *Instance) Seed() int64 {
	return i.value
}

func (i *Instance) djb2String(input string) {
	for _, c := range input {
		i.djb2Int64(int64(c))
	}
}

// super fast random algorithm.
// http://www.cse.yorku.ca/~oz/hash.html
func (i *Instance) djb2Int64(input int64) {
	i.value = (i.value << 5) + i.value + input
}

// AddEntropy uses various runtime variables to add entropy to the existing seed.
func (i *Instance) addEntropy() {
	// FIXME: reapply the 5381 initial value?

	// inherit previous entropy
	// nothing to do

	// handle callers
	{
		caller1 := std.GetCallerAt(1).String()
		i.djb2String(caller1)
		caller2 := std.GetCallerAt(2).String()
		i.djb2String(caller2)
	}

	// height
	{
		height := std.GetHeight()
		i.djb2Int64(height)
	}

	// time
	{
		secs := time.Now().Second()
		i.djb2Int64(int64(secs))
		nsecs := time.Now().Nanosecond()
		i.djb2Int64(int64(nsecs))
	}

	// FIXME: compute other hard-to-guess but deterministic variables, like real gas?
}

func (i *Instance) Value() int64 {
	i.addEntropy()
	return i.value
}
