// Xorshiftr128+ is a very fast psuedo-random number generation algorithm with strong
// statistical properties.
//
// The default random number algorithm in gno was ported from Go's v2 rand implementatoon, which
// defaults to the PCG algorithm. This algorithm is commonly used in language PRNG implementations
// because it has modest seeding requirements, and generates statistically strong randomness.
//
// This package provides an implementation of the Xorshiftr128+ PRNG algorithm. This algorithm provides
// strong statistical performance with most seeds (just don't seed it with zeros), and the performance
// of this implementation in Gno is more than four times faster than the default PCG implementation in
// `math/rand`.
//
//	Benchmark
//	---------
//	PCG:           1000000 Uint64 generated in 15.48s
//	Xorshiftr128+: 1000000 Uint64 generated in 3.22s
//	Ratio:         x4.81 times faster than PCG
//
// Use it directly:
//
//	prng = xorshiftr128plus.New() // pass a uint64 to seed it or pass nothing to seed it with entropy
//
// Or use it as a drop-in replacement for the default PRNT in Rand:
//
//	source = xorshiftr128plus.New()
//	prng := rand.New(source)
package xorshiftr128plus

import (
	"errors"
	"math"

	"gno.land/p/demo/entropy"
	"gno.land/p/nt/ufmt/v0"
)

type Xorshiftr128Plus struct {
	seed [2]uint64 // Seeds
}

func New(seeds ...uint64) *Xorshiftr128Plus {
	var s1, s2 uint64
	seed_length := len(seeds)
	if seed_length < 2 {
		e := entropy.New()
		if seed_length == 0 {
			s1 = e.Value64()
			s2 = e.Value64()
		} else {
			s1 = seeds[0]
			s2 = e.Value64()
		}
	} else {
		s1 = seeds[0]
		s2 = seeds[1]
	}

	prng := &Xorshiftr128Plus{}
	prng.Seed(s1, s2)
	return prng
}

func (x *Xorshiftr128Plus) Seed(s1, s2 uint64) {
	if s1 == 0 && s2 == 0 {
		panic("Seeds must not both be zero")
	}
	x.seed[0] = s1
	x.seed[1] = s2
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
var marshalXorshiftr128PlusLabel = []byte("xorshiftr128+:")

// MarshalBinary() returns a byte array that encodes the state of the PRNG. This can later be used
// with UnmarshalBinary() to restore the state of the PRNG.
// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (xs *Xorshiftr128Plus) MarshalBinary() ([]byte, error) {
	b := make([]byte, 30)
	copy(b, marshalXorshiftr128PlusLabel)
	bePutUint64(b[14:], xs.seed[0])
	bePutUint64(b[22:], xs.seed[1])
	return b, nil
}

// errUnmarshalXorshiftr128Plus is returned when unmarshalling fails.
var errUnmarshalXorshiftr128Plus = errors.New("invalid Xorshiftr128Plus encoding")

// UnmarshalBinary() restores the state of the PRNG from a byte array that was created with MarshalBinary().
// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (xs *Xorshiftr128Plus) UnmarshalBinary(data []byte) error {
	if len(data) != 30 || string(data[:14]) != string(marshalXorshiftr128PlusLabel) {
		return errUnmarshalXorshiftr128Plus
	}
	xs.seed[0] = beUint64(data[14:])
	xs.seed[1] = beUint64(data[22:])
	return nil
}

func (x *Xorshiftr128Plus) Uint64() uint64 {
	x0 := x.seed[0]
	x1 := x.seed[1]
	x.seed[0] = x1
	x0 ^= x0 << 23
	x0 ^= x0 >> 17
	x0 ^= x1
	x.seed[1] = x0 + x1
	return x.seed[1]
}

// Until there is better benchmarking support in gno, you can test the performance of this PRNG with this function.
// This isn't perfect, since it will include the startup time of gno in the results, but this will give you a timing
// for generating a million random uint64 numbers on any unix based system:
//
// `time gno run -expr 'benchmarkXorshiftr128Plus()' xorshiftr128plus.gno
func benchmarkXorshiftr128Plus(_iterations ...int) {
	iterations := 1000000
	if len(_iterations) > 0 {
		iterations = _iterations[0]
	}
	xs128p := New()

	for i := 0; i < iterations; i++ {
		_ = xs128p.Uint64()
	}
	ufmt.Println(ufmt.Sprintf("Xorshiftr128Plus: generate %d uint64\n", iterations))
}

// The averageXorshiftr128Plus() function is a simple benchmarking helper to demonstrate
// the most basic statistical property of the Xorshiftr128+ PRNG.
func averageXorshiftr128Plus(_iterations ...int) {
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
	xs128p := New()

	var average float64 = 0
	for i := 0; i < iterations; i++ {
		n := xs128p.Uint64()%(target*2) + 1
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

	ufmt.Println(ufmt.Sprintf("Xorshiftr128+ average of %d uint64: %f\n", iterations, average))
	ufmt.Println(ufmt.Sprintf("Xorshiftr128+ standard deviation  : %f\n", math.Sqrt(float64(sum_of_squares)/float64(iterations))))
	ufmt.Println(ufmt.Sprintf("Xorshiftr128+ theoretical perfect deviation: %f\n", (float64(target*2)-1)/math.Sqrt(12)))
}
