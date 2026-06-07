// This is a port of the 64-bit version of the ISAAC cryptographically secure PRNG, originally
// based on the reference implementation found at https://burtleburtle.net/bob/rand/isaacafa.html
//
// ISAAC has excellent statistical properties, with long cycle times, and uniformly distributed,
// unbiased, and unpredictable number generation. It can not be distinguished from real random
// data, and in three decades of scrutiny, no practical attacks have been found.
//
// The default random number algorithm in gno was ported from Go's v2 rand implementatoon, which
// defaults to the PCG algorithm. This algorithm is commonly used in language PRNG implementations
// because it has modest seeding requirements, and generates statistically strong randomness.
//
// This package provides an implementation of the 64-bit ISAAC PRNG algorithm. This algorithm
// provides very strong statistical performance, and is cryptographically secure, while still
// being substantially faster than the default PCG implementation in `math/rand`.
//
// Note that the approach to seeing with ISAAC is very important for best results, and seeding with
// ISAAC is not as simple as seeding with a single uint64 value. The ISAAC algorithm requires a
// 256-element seed. If used for cryptographic purposes, this will likely require entropy generated
// off-chain for actual cryptographically secure seeding. For other purposes, however, one can
// utilize the built-in seeding mechanism, which will leverage the xorshiftr128plus PRNG to generate
// any missing seeds if fewer than 256 are provided.
//
//		Benchmark
//		---------
//		PCG:         1000000 Uint64 generated in 15.58s
//		ISAAC:       1000000 Uint64 generated in 8.95s
//	        ISAAC:       1000000 Uint32 generated in 7.66s
//		Ratio:       x1.74 times faster than PCG (uint64)
//	        Ratio:       x2.03 times faster than PCG (uint32)
//
// Use it directly:
//
//		prng = isaac.New() // pass 0 to 256 uint64 seeds; if fewer than 256 are provided, the rest
//	                    // will be generated using the xorshiftr128plus PRNG.
//
// Or use it as a drop-in replacement for the default PRNT in Rand:
//
//	source = isaac64.New()
//	prng := rand.New(source)
package isaac64

import (
	"errors"
	"math"

	"gno.land/p/demo/entropy"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/wyhaines/rand/xorshiftr128plus"
)

const (
	RANDSIZL = 8
	RANDSIZ  = 1 << RANDSIZL // 256
)

type ISAAC struct {
	randrsl    [256]uint64
	randcnt    uint64
	mm         [256]uint64
	aa, bb, cc uint64
	seed       [256]uint64
}

// ISAAC requires a large, 256-element seed. This implementation will leverage the entropy
// package combined with the xorshiftr128plus PRNG to generate any missing seeds if fewer than
// the required number of arguments are provided.
func New(seeds ...uint64) *ISAAC {
	isaac := &ISAAC{}
	seed := [256]uint64{}

	index := 0
	for index = 0; index < len(seeds) && index < 256; index++ {
		seed[index] = seeds[index]
	}

	if index < 2 {
		e := entropy.New()
		for ; index < 2; index++ {
			seed[index] = e.Value64()
		}
	}

	// Use the first two seeds as seeding inputs for xorshiftr128plus, in order to
	// use it to provide any remaining missing seeds.
	prng := xorshiftr128plus.New(
		seed[0],
		seed[1],
	)
	for ; index < 256; index++ {
		seed[index] = prng.Uint64()
	}
	isaac.Seed(seed)
	return isaac
}

// Reinitialize the generator with a new seed. A seed must be composed of 256 uint64.
func (isaac *ISAAC) Seed(seed [256]uint64) {
	isaac.randrsl = seed
	isaac.seed = seed
	isaac.randinit(true)
}

// beUint64() decodes a uint64 from a set of eight bytes, assuming big endian encoding.
func beUint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler
	return uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
}

// bePutUint64() encodes a uint64 into a buffer of eight bytes.
func bePutUint64(b []byte, v uint64) {
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
}

// A label to identify the marshalled data.
var marshalISAACLabel = []byte("isaac:")

// MarshalBinary() returns a byte array that encodes the state of the PRNG. This can later be used
// with UnmarshalBinary() to restore the state of the PRNG.
// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (isaac *ISAAC) MarshalBinary() ([]byte, error) {
	b := make([]byte, 6+2048*3+8*3+8) // 6 + 2048*3 + 8*3 + 8 == 6182
	copy(b, marshalISAACLabel)
	offset := 6
	for i := 0; i < 256; i++ {
		bePutUint64(b[offset:], isaac.seed[i])
		offset += 8
	}
	for i := 0; i < 256; i++ {
		bePutUint64(b[offset:], isaac.randrsl[i])
		offset += 8
	}
	for i := 0; i < 256; i++ {
		bePutUint64(b[offset:], isaac.mm[i])
		offset += 8
	}
	bePutUint64(b[offset:], isaac.aa)
	offset += 8
	bePutUint64(b[offset:], isaac.bb)
	offset += 8
	bePutUint64(b[offset:], isaac.cc)
	offset += 8
	bePutUint64(b[offset:], isaac.randcnt)
	return b, nil
}

// errUnmarshalISAAC is returned when unmarshalling fails.
var errUnmarshalISAAC = errors.New("invalid ISAAC encoding")

// UnmarshalBinary() restores the state of the PRNG from a byte array that was created with MarshalBinary().
// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (isaac *ISAAC) UnmarshalBinary(data []byte) error {
	if len(data) != 6182 || string(data[:6]) != string(marshalISAACLabel) {
		return errUnmarshalISAAC
	}
	offset := 6
	for i := 0; i < 256; i++ {
		isaac.seed[i] = beUint64(data[offset:])
		offset += 8
	}
	for i := 0; i < 256; i++ {
		isaac.randrsl[i] = beUint64(data[offset:])
		offset += 8
	}
	for i := 0; i < 256; i++ {
		isaac.mm[i] = beUint64(data[offset:])
		offset += 8
	}
	isaac.aa = beUint64(data[offset:])
	offset += 8
	isaac.bb = beUint64(data[offset:])
	offset += 8
	isaac.cc = beUint64(data[offset:])
	offset += 8
	isaac.randcnt = beUint64(data[offset:])
	return nil
}

func (isaac *ISAAC) randinit(flag bool) {
	var a, b, c, d, e, f, g, h uint64
	isaac.aa = 0
	isaac.bb = 0
	isaac.cc = 0

	a = 0x9e3779b97f4a7c13
	b = 0x9e3779b97f4a7c13
	c = 0x9e3779b97f4a7c13
	d = 0x9e3779b97f4a7c13
	e = 0x9e3779b97f4a7c13
	f = 0x9e3779b97f4a7c13
	g = 0x9e3779b97f4a7c13
	h = 0x9e3779b97f4a7c13

	// scramble it
	for i := 0; i < 4; i++ {
		mix(&a, &b, &c, &d, &e, &f, &g, &h)
	}

	// fill in mm[] with messy stuff
	for i := 0; i < RANDSIZ; i += 8 {
		if flag {
			a += isaac.randrsl[i]
			b += isaac.randrsl[i+1]
			c += isaac.randrsl[i+2]
			d += isaac.randrsl[i+3]
			e += isaac.randrsl[i+4]
			f += isaac.randrsl[i+5]
			g += isaac.randrsl[i+6]
			h += isaac.randrsl[i+7]
		}
		mix(&a, &b, &c, &d, &e, &f, &g, &h)
		isaac.mm[i] = a
		isaac.mm[i+1] = b
		isaac.mm[i+2] = c
		isaac.mm[i+3] = d
		isaac.mm[i+4] = e
		isaac.mm[i+5] = f
		isaac.mm[i+6] = g
		isaac.mm[i+7] = h
	}

	if flag {
		// do a second pass to make all of the seed affect all of mm
		for i := 0; i < RANDSIZ; i += 8 {
			a += isaac.mm[i]
			b += isaac.mm[i+1]
			c += isaac.mm[i+2]
			d += isaac.mm[i+3]
			e += isaac.mm[i+4]
			f += isaac.mm[i+5]
			g += isaac.mm[i+6]
			h += isaac.mm[i+7]
			mix(&a, &b, &c, &d, &e, &f, &g, &h)
			isaac.mm[i] = a
			isaac.mm[i+1] = b
			isaac.mm[i+2] = c
			isaac.mm[i+3] = d
			isaac.mm[i+4] = e
			isaac.mm[i+5] = f
			isaac.mm[i+6] = g
			isaac.mm[i+7] = h
		}
	}

	isaac.isaac()
	isaac.randcnt = RANDSIZ
}

func mix(a, b, c, d, e, f, g, h *uint64) {
	*a -= *e
	*f ^= *h >> 9
	*h += *a

	*b -= *f
	*g ^= *a << 9
	*a += *b

	*c -= *g
	*h ^= *b >> 23
	*b += *c

	*d -= *h
	*a ^= *c << 15
	*c += *d

	*e -= *a
	*b ^= *d >> 14
	*d += *e

	*f -= *b
	*c ^= *e << 20
	*e += *f

	*g -= *c
	*d ^= *f >> 17
	*f += *g

	*h -= *d
	*e ^= *g << 14
	*g += *h
}

func ind(mm []uint64, x uint64) uint64 {
	return mm[(x>>3)&(RANDSIZ-1)]
}

func (isaac *ISAAC) isaac() {
	var a, b, x, y uint64
	a = isaac.aa
	b = isaac.bb + isaac.cc + 1
	isaac.cc++

	m := isaac.mm[:]
	r := isaac.randrsl[:]

	var i, m2Index int

	// First half
	for i = 0; i < RANDSIZ/2; i++ {
		m2Index = i + RANDSIZ/2
		switch i % 4 {
		case 0:
			a = ^(a ^ (a << 21)) + m[m2Index]
		case 1:
			a = (a ^ (a >> 5)) + m[m2Index]
		case 2:
			a = (a ^ (a << 12)) + m[m2Index]
		case 3:
			a = (a ^ (a >> 33)) + m[m2Index]
		}
		x = m[i]
		y = ind(m, x) + a + b
		m[i] = y
		b = ind(m, y>>RANDSIZL) + x
		r[i] = b
	}

	// Second half
	for i = RANDSIZ / 2; i < RANDSIZ; i++ {
		m2Index = i - RANDSIZ/2
		switch i % 4 {
		case 0:
			a = ^(a ^ (a << 21)) + m[m2Index]
		case 1:
			a = (a ^ (a >> 5)) + m[m2Index]
		case 2:
			a = (a ^ (a << 12)) + m[m2Index]
		case 3:
			a = (a ^ (a >> 33)) + m[m2Index]
		}
		x = m[i]
		y = ind(m, x) + a + b
		m[i] = y
		b = ind(m, y>>RANDSIZL) + x
		r[i] = b
	}

	isaac.bb = b
	isaac.aa = a
}

// Return a 64 bit random integer.
func (isaac *ISAAC) Uint64() uint64 {
	if isaac.randcnt == 0 {
		isaac.isaac()
		isaac.randcnt = RANDSIZ
	}
	isaac.randcnt--
	return isaac.randrsl[isaac.randcnt]
}

var gencycle int = 0
var bufferFor32 uint64 = uint64(0)

// Return a 32 bit random integer, composed of the high 32 bits of the generated 32 bit result.
func (isaac *ISAAC) Uint32() uint32 {
	if gencycle == 0 {
		bufferFor32 = isaac.Uint64()
		gencycle = 1
		return uint32(bufferFor32 >> 32)
	}

	gencycle = 0
	return uint32(bufferFor32 & 0xffffffff)
}

// Until there is better benchmarking support in gno, you can test the performance of this PRNG with this function.
// This isn't perfect, since it will include the startup time of gno in the results, but this will give you a timing
// for generating a million random uint64 numbers on any unix based system:
//
// `time gno run -expr 'benchmarkISAAC()' isaac64.gno
func benchmarkISAAC(_iterations ...int) {
	iterations := 1000000
	if len(_iterations) > 0 {
		iterations = _iterations[0]
	}
	isaac := New()

	for i := 0; i < iterations; i++ {
		_ = isaac.Uint64()
	}
	ufmt.Println(ufmt.Sprintf("ISAAC: generated %d uint64\n", iterations))
}

// The averageISAAC() function is a simple benchmarking helper to demonstrate
// the most basic statistical property of the ISAAC PRNG.
func averageISAAC(_iterations ...int) {
	target := uint64(500000)
	iterations := 1000000

	ufmt.Println(
		ufmt.Sprintf(
			"Averaging %d random numbers. The average should be very close to %d.\n",
			iterations,
			target))

	if len(_iterations) > 0 {
		iterations = _iterations[0]
	}
	isaac := New(987654321987654321, 123456789987654321, 1, 997755331886644220)

	var average float64 = 0
	var squares []uint64 = make([]uint64, iterations)
	for i := 0; i < iterations; i++ {
		n := isaac.Uint64()%(target*2) + 1
		average += (float64(n) - average) / float64(i+1)
		squares[i] = n
	}

	sum_of_squares := uint64(0)
	// transform numbers into their squares of the distance from the average
	for i := 0; i < iterations; i++ {
		difference := average - float64(squares[i])
		square := uint64(difference * difference)
		sum_of_squares += square
	}

	ufmt.Println(ufmt.Sprintf("ISAAC average of %d uint64: %f\n", iterations, average))
	ufmt.Println(ufmt.Sprintf("ISAAC standard deviation  : %f\n", math.Sqrt(float64(sum_of_squares)/float64(iterations))))
	ufmt.Println(ufmt.Sprintf("ISAAC theoretical perfect deviation: %f\n", (float64(target*2)-1)/math.Sqrt(12)))
}
