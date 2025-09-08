// This is a port of the ISAAC cryptographically secure PRNG, originally based on the reference
// implementation found at https://burtleburtle.net/bob/rand/isaacafa.html
//
// ISAAC has excellent statistical properties, with long cycle times, and uniformly distributed,
// unbiased, and unpredictable number generation. It can not be distinguished from real random
// data, and in three decades of scrutiny, no practical attacks have been found.
//
// The default random number algorithm in gno was ported from Go's v2 rand implementation, which
// defaults to the PCG algorithm. This algorithm is commonly used in language PRNG implementations
// because it has modest seeding requirements, and generates statistically strong randomness.
//
// This package provides an implementation of the 32-bit ISAAC PRNG algorithm. This
// algorithm provides very strong statistical performance, and is cryptographically
// secure, while still being substantially faster than the default PCG
// implementation in `math/rand`. Note that this package does implement a `Uint64()`
// function in order to generate a 64 bit number out of two 32 bit numbers. Doing this
// makes the generator only slightly faster than PCG, however,
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
//		ISAAC:       1000000 Uint64 generated in 13.23s
//		ISAAC:       1000000 Uint32 generated in 6.43s
//	     Ratio:       x1.18 times faster than PCG (uint64)
//	     Ratio:       x2.42 times faster than PCG (uint32)
//
// Use it directly:
//
//		prng = isaac.New() // pass 0 to 256 uint32 seeds; if fewer than 256 are provided, the rest
//	                    // will be generated using the xorshiftr128plus PRNG.
//
// Or use it as a drop-in replacement for the default PRNG in Rand:
//
//	source = isaac.New()
//	prng := rand.New(source)
package isaac

import (
	"errors"
	"math"
	"math/rand"

	"gno.land/p/demo/entropy"
	"gno.land/p/nt/ufmt"
	"gno.land/p/wyhaines/rand/xorshiftr128plus"
)

type ISAAC struct {
	randrsl    [256]uint32
	randcnt    uint32
	mm         [256]uint32
	aa, bb, cc uint32
	seed       [256]uint32
}

// ISAAC requires a large, 256-element seed. This implementation will leverage the entropy
// package combined with the the xorshiftr128plus PRNG to generate any missing seeds of
// fewer than the required number of arguments are provided.
func New(seeds ...uint32) *ISAAC {
	isaac := &ISAAC{}
	seed := [256]uint32{}

	index := 0
	for index = 0; index < len(seeds); index++ {
		seed[index] = seeds[index]
	}

	if index < 4 {
		e := entropy.New()
		for ; index < 4; index++ {
			seed[index] = e.Value()
		}
	}

	// Use up to the first four seeds as seeding inputs for xorshiftr128+, in order to
	// use it to provide any remaining missing seeds.
	prng := xorshiftr128plus.New(
		(uint64(seed[0])<<32)|uint64(seed[1]),
		(uint64(seed[2])<<32)|uint64(seed[3]),
	)
	for ; index < 256; index += 2 {
		val := prng.Uint64()
		seed[index] = uint32(val & 0xffffffff)
		if index+1 < 256 {
			seed[index+1] = uint32(val >> 32)
		}
	}
	isaac.Seed(seed)
	return isaac
}

func (isaac *ISAAC) Seed(seed [256]uint32) {
	isaac.randrsl = seed
	isaac.seed = seed
	isaac.randinit(true)
}

// beUint32() decodes a uint32 from a set of four bytes, assuming big endian encoding.
// binary.bigEndian.Uint32, copied to avoid dependency
func beUint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
}

// bePutUint32() encodes a uint64 into a buffer of eight bytes.
// binary.bigEndian.PutUint32, copied to avoid dependency
func bePutUint32(b []byte, v uint32) {
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}

// A label to identify the marshalled data.
var marshalISAACLabel = []byte("isaac:")

// MarshalBinary() returns a byte array that encodes the state of the PRNG. This can later be used
// with UnmarshalBinary() to restore the state of the PRNG.
// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (isaac *ISAAC) MarshalBinary() ([]byte, error) {
	b := make([]byte, 3094) // 6 + 1024 + 1024 + 1024 + 4 + 4 + 4 + 4 == 3090
	copy(b, marshalISAACLabel)
	for i := 0; i < 256; i++ {
		bePutUint32(b[6+i*4:], isaac.seed[i])
	}
	for i := 256; i < 512; i++ {
		bePutUint32(b[6+i*4:], isaac.randrsl[i-256])
	}
	for i := 512; i < 768; i++ {
		bePutUint32(b[6+i*4:], isaac.mm[i-512])
	}
	bePutUint32(b[3078:], isaac.aa)
	bePutUint32(b[3082:], isaac.bb)
	bePutUint32(b[3086:], isaac.cc)
	bePutUint32(b[3090:], isaac.randcnt)

	return b, nil
}

// errUnmarshalISAAC is returned when unmarshalling fails.
var errUnmarshalISAAC = errors.New("invalid ISAAC encoding")

// UnmarshalBinary() restores the state of the PRNG from a byte array that was created with MarshalBinary().
// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (isaac *ISAAC) UnmarshalBinary(data []byte) error {
	if len(data) != 3094 || string(data[:6]) != string(marshalISAACLabel) {
		return errUnmarshalISAAC
	}
	for i := 0; i < 256; i++ {
		isaac.seed[i] = beUint32(data[6+i*4:])
	}
	for i := 256; i < 512; i++ {
		isaac.randrsl[i-256] = beUint32(data[6+i*4:])
	}
	for i := 512; i < 768; i++ {
		isaac.mm[i-512] = beUint32(data[6+i*4:])
	}
	isaac.aa = beUint32(data[3078:])
	isaac.bb = beUint32(data[3082:])
	isaac.cc = beUint32(data[3086:])
	isaac.randcnt = beUint32(data[3090:])
	return nil
}

func (isaac *ISAAC) randinit(flag bool) {
	isaac.aa = 0
	isaac.bb = 0
	isaac.cc = 0

	var a, b, c, d, e, f, g, h uint32 = 0x9e3779b9, 0x9e3779b9, 0x9e3779b9, 0x9e3779b9, 0x9e3779b9, 0x9e3779b9, 0x9e3779b9, 0x9e3779b9

	for i := 0; i < 4; i++ {
		a ^= b << 11
		d += a
		b += c
		b ^= c >> 2
		e += b
		c += d
		c ^= d << 8
		f += c
		d += e
		d ^= e >> 16
		g += d
		e += f
		e ^= f << 10
		h += e
		f += g
		f ^= g >> 4
		a += f
		g += h
		g ^= h << 8
		b += g
		h += a
		h ^= a >> 9
		c += h
		a += b
	}

	for i := 0; i < 256; i += 8 {
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

		a ^= b << 11
		d += a
		b += c
		b ^= c >> 2
		e += b
		c += d
		c ^= d << 8
		f += c
		d += e
		d ^= e >> 16
		g += d
		e += f
		e ^= f << 10
		h += e
		f += g
		f ^= g >> 4
		a += f
		g += h
		g ^= h << 8
		b += g
		h += a
		h ^= a >> 9
		c += h
		a += b

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
		for i := 0; i < 256; i += 8 {
			a += isaac.mm[i]
			b += isaac.mm[i+1]
			c += isaac.mm[i+2]
			d += isaac.mm[i+3]
			e += isaac.mm[i+4]
			f += isaac.mm[i+5]
			g += isaac.mm[i+6]
			h += isaac.mm[i+7]

			a ^= b << 11
			d += a
			b += c
			b ^= c >> 2
			e += b
			c += d
			c ^= d << 8
			f += c
			d += e
			d ^= e >> 16
			g += d
			e += f
			e ^= f << 10
			h += e
			f += g
			f ^= g >> 4
			a += f
			g += h
			g ^= h << 8
			b += g
			h += a
			h ^= a >> 9
			c += h
			a += b

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
	isaac.randcnt = uint32(256)
}

func (isaac *ISAAC) isaac() {
	isaac.cc++
	isaac.bb += isaac.cc

	for i := 0; i < 256; i++ {
		x := isaac.mm[i]
		switch i % 4 {
		case 0:
			isaac.aa ^= isaac.aa << 13
		case 1:
			isaac.aa ^= isaac.aa >> 6
		case 2:
			isaac.aa ^= isaac.aa << 2
		case 3:
			isaac.aa ^= isaac.aa >> 16
		}
		isaac.aa += isaac.mm[(i+128)&0xff]

		y := isaac.mm[(x>>2)&0xff] + isaac.aa + isaac.bb
		isaac.mm[i] = y
		isaac.bb = isaac.mm[(y>>10)&0xff] + x
		isaac.randrsl[i] = isaac.bb
	}
}

// Returns a random uint32.
func (isaac *ISAAC) Uint32() uint32 {
	if isaac.randcnt == uint32(0) {
		isaac.isaac()
		isaac.randcnt = uint32(256)
	}
	isaac.randcnt--
	return isaac.randrsl[isaac.randcnt]
}

// Returns a random uint64 by combining two uint32s.
func (isaac *ISAAC) Uint64() uint64 {
	return uint64(isaac.Uint32()) | (uint64(isaac.Uint32()) << 32)
}

// Until there is better benchmarking support in gno, you can test the performance of this PRNG with this function.
// This isn't perfect, since it will include the startup time of gno in the results, but this will give you a timing
// for generating a million random uint64 numbers on any unix based system:
//
// `time gno run -expr 'benchmarkISAAC()' xorshift64star.gno
func benchmarkISAAC(_iterations ...int) {
	iterations := 1000000
	if len(_iterations) > 0 {
		iterations = _iterations[0]
	}
	isaac := New()

	for i := 0; i < iterations; i++ {
		_ = isaac.Uint64()
	}
	ufmt.Println(ufmt.Sprintf("ISAAC: generate %d uint64\n", iterations))
}

// The averageISAAC() function is a simple benchmarking helper to demonstrate
// the most basic statistical property of the ISAAC PRNG.
func averageISAAC(_iterations ...int) {
	target := uint64(500000)
	iterations := 1000000
	var squares [1000000]uint64

	ufmt.Println(
		ufmt.Sprintf(
			"Averaging %d random numbers. The average should be very close to %d.\n",
			iterations,
			target))

	if len(_iterations) > 0 {
		iterations = _iterations[0]
	}
	isaac := New(987654321, 123456789, 999999999, 111111111)

	var average float64 = 0
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

func averagePCG(_iterations ...int) {
	target := uint64(500000)
	iterations := 1000000
	var squares [1000000]uint64

	ufmt.Println(
		ufmt.Sprintf(
			"Averaging %d random numbers. The average should be very close to %d.\n",
			iterations,
			target))

	if len(_iterations) > 0 {
		iterations = _iterations[0]
	}
	isaac := rand.NewPCG(987654321, 123456789)

	var average float64 = 0
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

	ufmt.Println(ufmt.Sprintf("PCG average of %d uint64: %f\n", iterations, average))
	ufmt.Println(ufmt.Sprintf("PCG standard deviation  : %f\n", math.Sqrt(float64(sum_of_squares)/float64(iterations))))
	ufmt.Println(ufmt.Sprintf("PCG theoretical perfect deviation: %f\n", (float64(target*2)-1)/math.Sqrt(12)))
}
