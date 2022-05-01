package rand

// the goal of this package is to implement a random library that
// is fully deterministic for validators while being hard to
// determine by the callers.

import (
	"crypto/md5"
	"encoding/binary"
	"math/rand"
	"std"
)

type instance struct {
	seed int64
}

func New() *instance {
	r := instance{}
	r.addEntropy()
	return &r
}

// AddEntropy uses various runtime variables to add entropy to the existing seed.
func (i *instance) addEntropy() {
	h := md5.New()

	// why not?
	h.Write([]byte("now, you gno"))

	// inherit previous entropy
	{
		buf := make([]byte, 16)
		_, _ = rand.New(rand.NewSource(i.seed)).Read(buf)
		h.Write(buf)
	}

	// callers
	h.Write([]byte(std.GetCallerAt(1)))
	h.Write([]byte(std.GetCallerAt(2)))

	// height
	{
		height := std.GetHeight()
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(height))
		h.Write(buf)
	}

	// time
	{
		now := std.GetTimestamp()
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(now))
		h.Write(buf)
	}

	// set seed
	i.seed = int64(binary.BigEndian.Uint64(h.Sum(nil)))
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
