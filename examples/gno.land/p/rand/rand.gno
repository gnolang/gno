package rand

// the goal of this package is to implement a random library that
// is fully deterministic for validators while being hard to
// determine by the callers.
//
// We use the Bernstein's hash djb2 to be CPU-cycle efficient.

import (
	"math/rand"
	"std"
)

type instance struct {
	seed int64
}

func New() *instance {
	r := instance{seed: 5381}
	r.addEntropy()
	return &r
}

// http://www.cse.yorku.ca/~oz/hash.html
func (i *instance) djb2String(input string) {
	for _, c := range input {
		i.djb2Int64(int64(c))
	}
}

func (i *instance) djb2Int64(input int64) {
	i.seed = (i.seed << 5) + i.seed + input
}

// AddEntropy uses various runtime variables to add entropy to the existing seed.
func (i *instance) addEntropy() {
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
		now := std.GetTimestamp()
		i.djb2Int64(int64(now))
	}

	// FIXME: compute other hard-to-guess but deterministic variables, like real gas?
}

func (i *instance) Float32() float32 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Float32()
}

func (i *instance) Float64() float64 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Float64()
}

func (i *instance) Int() int {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Int()
}

func (i *instance) Intn(n int) int {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Intn(n)
}

func (i *instance) Int63() int64 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Int63()
}

func (i *instance) Int63n(n int64) int64 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Int63n(n)
}

func (i *instance) Int31() int32 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Int31()
}

func (i *instance) Int31n(n int32) int32 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Int31n(n)
}

func (i *instance) Uint32() uint32 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Uint32()
}

func (i *instance) Uint64() uint64 {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Uint64()
}

func (i *instance) Read(p []byte) (n int, err error) {
	i.addEntropy()
	return rand.New(rand.NewSource(i.seed)).Read(p)
}

func (i *instance) Shuffle(n int, swap func(i, j int)) {
	i.addEntropy()
	rand.New(rand.NewSource(i.seed)).Shuffle(n, swap)
}
