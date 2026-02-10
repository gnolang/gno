// Xorshift64* is a very fast psuedo-random number generation algorithm with strong
// statistical properties.
//
// The default random number algorithm in gno was ported from Go's v2 rand implementatoon, which
// defaults to the PCG algorithm. This algorithm is commonly used in language PRNG implementations
// because it has modest seeding requirements, and generates statistically strong randomness.
//
// This package provides an implementation of the Xorshift64* PRNG algorithm. This algorithm provides
// strong statistical performance with most seeds (just don't seed it with zero), and the performance
// of this implementation in Gno is more than four times faster than the default PCG implementation in
// `math/rand`.
//
//	Benchmark
//	---------
//	PCG:         1000000 Uint64 generated in 15.58s
//	Xorshift64*: 1000000 Uint64 generated in 3.77s
//	Ratio:       x4.11 times faster than PCG
//
// Use it directly:
//
//	prng = xorshift64star.New() // pass a uint64 to seed it or pass nothing to seed it with entropy
//
// Or use it as a drop-in replacement for the default PRNT in Rand:
//
//	source = xorshift64star.New()
//	prng := rand.New(source)
package xorshift64star

import (
	"errors"
	"math"

	"gno.land/p/demo/entropy"
	"gno.land/p/nt/ufmt"
)

// Xorshift64Star is a PRNG that implements the Xorshift64* algorithm.
type Xorshift64Star struct {
	seed uint64
}

// New() creates a new instance of the PRNG with a given seed, which
// should be a uint64. If no seed is provided, the PRNG will be seeded via the
// gno.land/p/demo/entropy package.
func New(seed ...uint64) *Xorshift64Star {
	xs := &Xorshift64Star{}
	xs.Seed(seed...)
	return xs
}

// Seed() implements the rand.Source interface. It provides a way to set the seed for the PRNG.
func (xs *Xorshift64Star) Seed(seed ...uint64) {
	if len(seed) == 0 {
		e := entropy.New()
		xs.seed = e.Value64()
	} else {
		xs.seed = seed[0]
	}
}

// beUint64() decodes a uint64 from a set of eight bytes, assuming big endian encoding.
// binary.bigEndian.Uint64, copied to avoid dependency
func beUint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
}

// bePutUint64() encodes a uint64 into a buffer of eight bytes.
// binary.bigEndian.PutUint64, copied to avoid dependency
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
var marshalXorshift64StarLabel = []byte("xorshift64*:")

// MarshalBinary() returns a byte array that encodes the state of the PRNG. This can later be used
// with UnmarshalBinary() to restore the state of the PRNG.
// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (xs *Xorshift64Star) MarshalBinary() ([]byte, error) {
	b := make([]byte, 20)
	copy(b, marshalXorshift64StarLabel)
	bePutUint64(b[12:], xs.seed)
	return b, nil
}

// errUnmarshalXorshift64Star is returned when unmarshalling fails.
var errUnmarshalXorshift64Star = errors.New("invalid Xorshift64* encoding")

// UnmarshalBinary() restores the state of the PRNG from a byte array that was created with MarshalBinary().
// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (xs *Xorshift64Star) UnmarshalBinary(data []byte) error {
	if len(data) != 20 || string(data[:12]) != string(marshalXorshift64StarLabel) {
		return errUnmarshalXorshift64Star
	}
	xs.seed = beUint64(data[12:])
	return nil
}

// Uint64() generates the next random uint64 value.
func (xs *Xorshift64Star) Uint64() uint64 {
	xs.seed ^= xs.seed >> 12
	xs.seed ^= xs.seed << 25
	xs.seed ^= xs.seed >> 27
	xs.seed *= 2685821657736338717
	return xs.seed // Operations naturally wrap around in uint64
}

// Until there is better benchmarking support in gno, you can test the performance of this PRNG with this function.
// This isn't perfect, since it will include the startup time of gno in the results, but this will give you a timing
// for generating a million random uint64 numbers on any unix based system:
//
// `time gno run -expr 'benchmarkXorshift64Star()' xorshift64star.gno
func benchmarkXorshift64Star(_iterations ...int) {
	iterations := 1000000
	if len(_iterations) > 0 {
		iterations = _iterations[0]
	}
	xs64s := New()

	for i := 0; i < iterations; i++ {
		_ = xs64s.Uint64()
	}
	ufmt.Println(ufmt.Sprintf("Xorshift64*: generate %d uint64\n", iterations))
}

// The averageXorshift64Star() function is a simple benchmarking helper to demonstrate
// the most basic statistical property of the Xorshift64* PRNG.
func averageXorshift64Star(_iterations ...int) {
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
	xs64s := New()

	var average float64 = 0
	for i := 0; i < iterations; i++ {
		n := xs64s.Uint64()%(target*2) + 1
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

	ufmt.Println(ufmt.Sprintf("Xorshift64* average of %d uint64: %f\n", iterations, average))
	ufmt.Println(ufmt.Sprintf("Xorshift64* standard deviation  : %f\n", math.Sqrt(float64(sum_of_squares)/float64(iterations))))
	ufmt.Println(ufmt.Sprintf("Xorshift64* theoretical perfect deviation: %f\n", (float64(target*2)-1)/math.Sqrt(12)))
}
