// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fuzzing

import (
	"fmt"
	"math"
	"math/bits"
)

// PCG is a PCG generator with 128 bits of internal state.
type PCG struct {
	hi uint64
	lo uint64
}

// used to calculate execution times; only present in testing stdlibs
func unixNano() int64

func UnixNano() int64 {
	return unixNano()
}

// NewPCG returns a new PCG seeded with the given two 64-bit values.
func NewPCG(seed1, seed2 uint64) *PCG {
	return &PCG{hi: seed1, lo: seed2}
}

// NewPCGFromInt64 returns a new PCG seeded with a single int64 value.
func NewPCGFromInt64(seed int64) *PCG {
	return &PCG{hi: 0, lo: uint64(seed)}
}

// Seed resets the PCG state.
func (p *PCG) Seed(seed1, seed2 uint64) {
	p.hi = seed1
	p.lo = seed2
}

// next() updates the internal state and performs a 128-bit multiplication
// and addition to generate a new (hi, lo) value.
func (p *PCG) next() (hi, lo uint64) {
	const (
		mulHi = 2549297995355413924
		mulLo = 4865540595714422341
		incHi = 6364136223846793005
		incLo = 1442695040888963407
	)
	hi, lo = bits.Mul64(p.lo, mulLo)
	hi += p.hi*mulLo + p.lo*mulHi
	lo, c := bits.Add64(lo, incLo, 0)
	hi, _ = bits.Add64(hi, incHi, c)
	p.lo = lo
	p.hi = hi
	return hi, lo
}

// Uint64 returns a uniformly distributed random uint64 value.
// Internally, it applies PCG-XSH-RR post-processing and DXSM (Double Xorshift Multiply).

func (p *PCG) Uint64() uint64 {
	hi, lo := p.next()
	const cheapMul = 0xda942042e4dd58b5
	hi ^= hi >> 32
	hi *= cheapMul
	hi ^= hi >> 48
	hi *= (lo | 1)
	return hi
}

// Int generates a random integer of type int.
// It removes the highest bit using bit masking to ensure a non-negative value.
func (p *PCG) Int() int {
	return int(p.Uint64() &^ (1 << 63)) // Removes the highest bit to return a non-negative value.
}

func (p *PCG) Intn(n int) int {
	if n <= 0 {
		panic("n must be greater than 0")
	}
	return int(p.Uint64() % uint64(n))
}

// Int32 generates a random 32-bit integer of type int32.
// It extracts the upper 32 bits from a 64-bit random value.
func (p *PCG) Int32() int32 {
	return int32(p.Uint64() >> 32) // Extracts the upper 32 bits and converts to int32
}

func (p *PCG) Int64() int64 {
	return int64(p.Uint64() &^ (1 << 63)) // Removes the highest bit to return a non-negative value.
}

// Uint32 returns a 32-bit random number.
func (p *PCG) Uint32() uint32 {
	return uint32(p.Uint64() >> 32)
}

// Uint64n returns a pseudo-random uint64 in the range [0, n).
func (p *PCG) Uint64n(n uint64) uint64 {
	if n == 0 {
		panic("n must be greater than 0")
	}
	return p.Uint64() % n
}

// Float64 returns a random float64 in [0.0, 1.0) using 53 bits of precision.
func (p *PCG) Float64() float64 {
	return float64(p.Uint64()>>11) / float64(1<<53)
}

// NormFloat64 returns a normally distributed random float64 (mean=0, stddev=1)
// using the Box–Muller transform.
func (p *PCG) NormFloat64() float64 {
	u1 := p.Float64()
	if u1 < 1e-12 {
		u1 = 1e-12
	}
	u2 := p.Float64()
	radius := math.Sqrt(-2 * math.Log(u1))
	theta := 2 * math.Pi * u2
	return radius * math.Cos(theta)
}

// RandRange returns a random int64 in [start, end).
func (p *PCG) RandRange(start, end int64) int64 {
	if start >= end {
		panic("start must be < end")
	}
	return start + int64(p.Uint64()%(uint64(end-start)))
}

// UintRandRange returns a random uint64 in [start, end).
func (p *PCG) UintRandRange(start, end uint64) uint64 {
	if start >= end {
		panic("start must be < end")
	}
	return start + p.Uint64()%(end-start)
}

// UniformRandomBool returns true with the given probability in [0,1].
func (p *PCG) UniformRandomBool(probability float64) bool {
	if probability < 0 || probability > 1 {
		panic("probability must be between 0 and 1")
	}
	return p.Float64() < probability
}

// GenerateRandomBool returns a boolean value using a bias based on a normal distribution.
func (p *PCG) GenerateRandomBool(bias float64) bool {
	if bias < 0 || bias > 1 {
		panic("bias must be in range [0,1]")
	}
	res := (p.NormFloat64() + 1) / 2
	return res > bias
}

var (
	singlePCG   *PCG
	initialized bool
)

// OncePCG returns a singleton instance of the PCG-based random generator.
func OncePCG() *PCG {
	if !initialized {
		initialized = true
		seed := uint64(unixNano())
		singlePCG = NewPCGFromInt64(int64(seed))
	}
	return singlePCG
}

// Nrand returns a value approximating a normal distribution scaled to roughly [-1,1].
func Nrand() float64 {
	pcg := OncePCG()
	val := pcg.NormFloat64()
	scaled := val / 3
	if scaled < -1 {
		return -1
	} else if scaled > 1 {
		return 1
	}
	return scaled
}

// RandRange returns a pseudo-random int64 in [start, end).
func RandRange(start, end int64) int64 {
	if start >= end {
		panic("start >= end")
	}
	pcg := OncePCG()
	// Int64() Effect: Uses a 63-bit value by removing the highest bit.

	randNum := int64(pcg.Uint64() &^ (1 << 63))
	if randNum < 0 {
		randNum = -randNum
	}
	hashedNum := randNum % (end - start)
	return start + hashedNum
}

// UintRandRange returns a pseudo-random uint64 in [start, end).
func UintRandRange(start, end uint64) uint64 {
	if start >= end {
		panic("start >= end")
	}
	pcg := OncePCG()
	randNum := pcg.Uint64()
	hashedNum := randNum % (end - start)
	return start + hashedNum
}

// RandInt64 returns a pseudo-random int64 (full range).
func RandInt64() int64 {
	pcg := OncePCG()
	return int64(pcg.Uint64() &^ (1 << 63))
}

// RandUint64 returns a pseudo-random uint64 (full range).
func RandUint64() uint64 {
	return OncePCG().Uint64()
}

// RandUint32 returns a pseudo-random uint32 (full range).
func RandUint32() uint32 {
	return OncePCG().Uint32()
}

// GenerateRandomBool returns a boolean value based on a bias using a normal distribution.
// Since nrand() has a range of [-1,1], it is converted to [0,1] before comparing with bias.
func GenerateRandomBool(bias float64) bool {
	if bias < 0 || bias > 1 {
		panic("bias should be in the range [0,1]")
	}
	res := (Nrand() + 1) / 2
	return res > bias
}

// UniformRandomBool returns true with the given probability.
func UniformRandomBool(probability float64) bool {
	if probability < 0.0 || probability > 1.0 {
		panic("Probability must be between 0.0 and 1.0")
	}
	return OncePCG().Float64() < probability
}
