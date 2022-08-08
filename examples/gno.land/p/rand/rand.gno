package rand

// Disclaimer: this package is unsafe and won't prevent others to
//             guess values in advance.
//
// the goal of this package is to implement a random library that
// is fully deterministic for validators while being hard to guess.
//
// We use the Bernstein's hash djb2 to be CPU-cycle efficient.

import (
	"math/rand"
	"std"
	"time"
)

type Instance struct {
	seed int64
}

func New() *Instance {
	r := Instance{seed: 5381}
	r.addEntropy()
	return &r
}

func FromSeed(seed int64) *Instance {
	r := Instance{seed: seed}
	r.addEntropy()
	return &r
}

func (i *Instance) Seed() int64 {
	return i.seed
}

func (i *Instance) djb2String(input string) {
	for _, c := range input {
		i.djb2Int64(int64(c))
	}
}

// super fast random algorithm.
// http://www.cse.yorku.ca/~oz/hash.html
func (i *Instance) djb2Int64(input int64) {
	i.seed = (i.seed << 5) + i.seed + input
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

func (i *Instance) Float32() float32 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Float32()
}

func (i *Instance) Float64() float64 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Float64()
}

func (i *Instance) Int() int {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Int()
}

func (i *Instance) Intn(n int) int {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Intn(n)
}

func (i *Instance) Int63() int64 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Int63()
}

func (i *Instance) Int63n(n int64) int64 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Int63n(n)
}

func (i *Instance) Int31() int32 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Int31()
}

func (i *Instance) Int31n(n int32) int32 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Int31n(n)
}

func (i *Instance) Uint32() uint32 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Uint32()
}

func (i *Instance) Uint64() uint64 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Uint64()
}

func (i *Instance) Read(p []byte) (n int, err error) {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Read(p)
}

func (i *Instance) Shuffle(n int, swap func(i, j int)) {
	i.addEntropy()
	rand.New(rand.NewSource(i.seed)).Shuffle(n, swap)
}
