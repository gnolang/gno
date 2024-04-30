package pcg

import (
	"encoding/binary"
	"errors"
	"math"
	"math/bits"
)

const (
	inv52 = 1.0 / (1 << 52)
	// inv64 = 1.0 / float64(math.MaxUint64)
)

// A PCG64 is a PCG64 generator with 128 bits of internal state.
// A zero PCG64 is equivalent to one seeded with 0.
type PCG64 struct {
	hi, lo *PCG32
}

// NewPCG64 returns a new PCG64 generator seeded with thr given values.
// seed1 and seed2 are the initial state values for the generator.
func NewPCG64(seed1, seed2 uint64) *PCG64 {
	return &PCG64{
		hi: NewPCG32().Seed(seed1, 0),
		lo: NewPCG32().Seed(seed2, 0),
	}
}

// Seed initializes the PCG64 generator with the given state and sequence values.
// seed1 and seed2 are the initial state values, and seq1 and seq2 are the sequence values.
func (p *PCG64) Seed(seed1, seed2, seq1, seq2 uint64) *PCG64 {
	mask := ^uint64(0) >> 1
	if seq1&mask == seq2&mask {
		seq2 = ^seq2
	}
	p.lo.Seed(seed1, seq1)
	p.hi.Seed(seed2, seq2)

	return p
}

// Uint64 generates a pseudorandom 64-bit unsigned integer using the PCG64 algorithm.
func (p *PCG64) Uint64() uint64 {
	return uint64(p.hi.Uint32())<<32 | uint64(p.lo.Uint32())
}

// Uint63 generates a pseudorandom 63-bit integer using the PCG64 algorithm.
// It masks the highest bit to ensure the value is within the 63-bit integer range.
func (p *PCG64) Uint63() int64 {
	return int64(p.Uint64() & 0x7FFFFFFFFFFFFFFF) // Mask the highest bit to stay within the 63-bit range
}

// Uint64n generates a pseudorandom number in the range [0, bound) using the PCG64 algorithm.
func (p *PCG64) Uint64n(bound uint64) uint64 {
	threshold := -bound % bound
	for {
		r := p.Uint64()
		if r >= threshold {
			return r % bound
		}
	}
}

// Float64 returns a random float64 in the range [0.0, 1.0).
func (p *PCG64) Float64() float64 {
	return float64(p.Uint63()>>11) * inv52
}

// Float64Full uses the full 64 bits of the generated number to produce a random float64.
// slightly more precise than Float64() but slower.
func (p *PCG64) Float64Full() float64 {
	return float64(p.Uint64()&0xFFFFFFFFFFFFFF) / math.MaxUint64
}

// Advance moves the PCG64 generator forward by `delta` steps.
// It updates the initial state of the generator.
func (p *PCG64) Advance(delta uint64) *PCG64 {
	p.hi.Advance(delta)
	p.lo.Advance(delta)
	return p
}

// Retreat moves the PCG64 generator backward by `delta` steps.
// it updates the initial state of the generator.
func (p *PCG64) Retreat(delta uint64) *PCG64 {
	safeDelta := ^uint64(0) - 1
	p.Advance(safeDelta)
	return p
}

func (p *PCG64) Shuffle(n int, swap func(i, j int)) {
	// Fisher-Yates shuffle: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
	for i := n - 1; i > 0; i-- {
		j := int(p.Uint64n(uint64(i + 1)))
		swap(i, j)
	}
}

func (p *PCG64) Perm(n int) []int {
	res := make([]int, n)
	for i := range res {
		res[i] = i
	}
	p.Shuffle(n, func(i, j int) {
		res[i], res[j] = res[j], res[i]
	})
	return res
}

// Read generates random bytes in the provided byte slice using the PCG64 random number generator.
// It employs loop unrolling to process 16 bytes at a time for performance enhancement.
func (p *PCG64) Read(buf []byte) (int, error) {
	n := len(buf)
	i := 0

	// Loop unrolling: processing 16 bytes per iteration
	for ; i <= n-16; i += 16 {
		val1 := p.Uint64()
		val2 := p.Uint64()
		binary.LittleEndian.PutUint64(buf[i:], val1)
		binary.LittleEndian.PutUint64(buf[i+8:], val2)
	}

	// Handle any remaining bytes that were not processed in the main loop
	if i < n {
		remaining := buf[i:]
		for j := 0; j < len(remaining); j += 8 {
			if i+j < n {
				val := p.Uint64()
				// Only write the necessary bytes
				for k := 0; k < 8 && (j+k) < len(remaining); k++ {
					remaining[j+k] = byte(val >> (8 * k))
				}
			}
		}
	}

	return n, nil
}

func beUint64(b []byte) uint64 {
	_ = b[7]
	return uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
}

func bePutUint64(b []byte, v uint64) {
	_ = b[7]
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
}

// MarshalBinaryPCG64 serializes the state of the PCG64 generator to a binary format.
// It returns the serialized state as a byte slice.
func (p *PCG64) MarshalBinaryPCG64() ([]byte, error) {
	b := make([]byte, 20)
	copy(b, "pcg:")
	bePutUint64(b[4:], p.hi.state)
	bePutUint64(b[4+8:], p.lo.state)
	return b, nil
}

var errUnmarshalPCG = errors.New("invalid PCG encoding")

// UnmarshalBinaryPCG64 deserializes the state of the PCG64 generator from a binary format.
// It takes the serialized state as a byte slice and updates the generator's state.
func (p *PCG64) UnmarshalBinary(b []byte) error {
	if len(b) != 20 || string(b[:4]) != "pcg:" {
		return errUnmarshalPCG
	}
	p.hi.state = beUint64(b[4:])
	p.lo.state = beUint64(b[4+8:])
	return nil
}

func (p *PCG64) next() (uint64, uint64) {
	const (
		mulHi = 2549297995355413924
		mulLo = 4865540595714422341
		incHi = 6364136223846793005
		incLo = 1442695040888963407
	)

	// state = state * mul + inc
	hi, lo := bits.Mul64(p.lo.state, mulLo)
	hi += p.hi.state*mulLo + p.lo.state*mulHi

	lo, c := bits.Add64(lo, incLo, 0)
	hi, _ = bits.Add64(hi, incHi, c)

	p.lo.state = lo
	p.hi.state = hi

	return hi, lo
}

// NextUInt64WithMCG generates a pseudorandom 64-bit unsigned integer using the PCG64 algorithm with Multiplier Congruential Generator (MCG).
// It updates the internal state of the generator and returns the generated value.
func (p *PCG64) Uint64nWithMCG() uint64 {
	hi, lo := p.next()

	// ref: https://www.pcg-random.org/posts/128-bit-mcg-passes-practrand.html (#64-bit Multiplier)
	const cheapMul = 0xda942042e4dd58b5 // 15750249268501108917
	hi ^= hi >> 22
	hi *= cheapMul
	hi ^= hi >> 48
	hi *= (lo | 1)

	return hi
}
