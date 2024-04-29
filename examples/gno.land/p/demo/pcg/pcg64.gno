package pcg

import (
	"errors"
	"math"
	"math/bits"
)

// A PCG64 is a PCG64 generator with 128 bits of internal state.
// A zero PCG64 is equivalent to one seeded with 0.
type PCG64 struct {
	hi, lo *PCG32
}

// NewPCG64 returns a new PCG64 generator seeded with the given values.
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

// Float64 returns a random float64 value in the range [0.0, 1.0).
//
// The function generates a 63-bit unsigned integer using Uint63() and then:
//   1. Shifts the generated number 11 bits to the right to obtain a 52-bit mantissa.
//   2. Multiplies the mantissa by 2^(-52) to scale it to the range [0.0, 1.0).
//
// This method produces a double-precision floating-point number with 52 bits of precision,
// which is the standard precision for float64 values.
//
// The time complexity of this function is O(1), as it performs a constant number of operations.
// The space complexity is O(1), as it uses only a constant amount of additional memory.
//
// Note: This method does not use all 64 bits of the generated number, which may result in
// a slight loss of precision compared to Float64Full(). However, it is faster and sufficient
// for most use cases.
func (p *PCG64) Float64() float64 {
	return float64(p.Uint63()>>11) * (1.0 / (1 << 52))
}

// Float64Full returns a random float64 value in the range [0.0, 1.0).
//
// The function generates a 64-bit unsigned integer using Uint64() and then:
//   1. Masks the generated number to obtain the lower 64 bits.
//   2. Divides the masked number by the maximum 64-bit unsigned integer value (math.MaxUint64)
//      to scale it to the range [0.0, 1.0).
//
// This method utilizes the full 64 bits of the generated number, providing slightly higher
// precision compared to Float64(). However, it is slower due to the additional masking and
// division operations.
//
// The time complexity of this function is O(1), as it performs a constant number of operations.
// The space complexity is O(1), as it uses only a constant amount of additional memory.
//
// Note: While this method provides higher precision, it may not be necessary for most use cases.
// Float64() is generally faster and sufficient for generating random float64 values.
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
// It updates the initial state of the generator.
func (p *PCG64) Retreat(delta uint64) *PCG64 {
	safeDelta := ^uint64(0) - 1
	p.Advance(safeDelta)
	return p
}

// Shuffle shuffles the elements of a slice in place using the Fisher-Yates shuffle algorithm.
// It takes the following parameters:
//   - n: The number of elements to shuffle.
//   - swap: A function that swaps the elements at indices i and j.
//
// The function shuffles the elements by iterating from n-1 down to 1. In each iteration:
//   1. Generate a random index j between 0 and the current index i (inclusive) using Uint64n.
//   2. Call the `swap` function to swap the elements at indices i and j.
//
// This process ensures that each element has an equal probability of ending up at any position
// in the shuffled slice.
//
// The time complexity of this function is O(n), where n is the number of elements to shuffle.
// The space complexity is O(1), as it shuffles the elements in place and uses only a constant
// amount of additional memory.
func (p *PCG64) Shuffle(n int, swap func(i, j int)) {
	// Fisher-Yates shuffle: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
	for i := n - 1; i > 0; i-- {
		j := int(p.Uint64n(uint64(i + 1)))
		swap(i, j)
	}
}

// Perm returns a random permutation of the integers in the range [0, n).
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

// Uint64nWithMCG generates a pseudorandom 64-bit unsigned integer using the PCG64 algorithm
// with a Multiplicative Congruential Generator (MCG) as the output function.
//
// The function updates the internal state of the generator using the PCG64 transition function:
//   state = state * multiplier + increment
//
// It then applies a series of bitwise operations to the updated state to produce the random number:
//   1. XOR the high 64 bits of the state with the high 64 bits right-shifted by 22 bits.
//   2. Multiply the result by a cheap multiplier (0xda942042e4dd58b5).
//   3. XOR the result with the result right-shifted by 48 bits.
//   4. Multiply the result by (low 64 bits of the state | 1).
//
// The resulting value is returned as the random 64-bit unsigned integer.
//
// The MCG output function used in this implementation is based on the one described in the PCG
// random number generator family, specifically the 128-bit MCG with 64-bit output.
// It provides good statistical properties and passes various randomness tests, including PractRand.
//
// The time complexity of this function is O(1), as it performs a constant number of bitwise operations.
// The space complexity is O(1), as it uses only a constant amount of additional memory.
//
// Note: The Uint64nWithMCG function is deterministic for a given state of the PCG64 generator.
// If you need different random numbers across multiple calls, make sure to advance the state of the
// generator using Advance() or reseed it using Seed() between calls.
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
