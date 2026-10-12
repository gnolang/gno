# package xorshift64star // import "gno.land/p/demo/math/rand/xorshift64star"

Xorshift64* is a very fast psuedo-random number generation algorithm with strong
statistical properties.

The default random number algorithm in gno was ported from Go's v2 rand
implementatoon, which defaults to the PCG algorithm. This algorithm is
commonly used in language PRNG implementations because it has modest seeding
requirements, and generates statistically strong randomness.

This package provides an implementation of the Xorshift64* PRNG algorithm.
This algorithm provides strong statistical performance with most seeds (just
don't seed it with zero), and the performance of this implementation in Gno is
more than four times faster than the default PCG implementation in `math/rand`.


```
Benchmark
---------
PCG:         1000000 Uint64 generated in 15.58s
Xorshift64*: 1000000 Uint64 generated in 3.77s
Ratio:       x4.11 times faster than PCG
```

Use it directly:

```
prng = xorshift64star.New() // pass a uint64 to seed it or pass nothing to seed it with entropy
```

Or use it as a drop-in replacement for the default PRNT in Rand:

```
source = xorshift64star.New()
prng := rand.New(source)
```

## TYPES

```
type Xorshift64Star struct {
	// Has unexported fields.
}
```

Xorshift64Star is a PRNG that implements the Xorshift64* algorithm.

`func New(seed ...uint64) *Xorshift64Star`
    New() creates a new instance of the PRNG with a given seed, which should
    be a uint64. If no seed is provided, the PRNG will be seeded via the
    gno.land/p/demo/entropy package.

`func (xs *Xorshift64Star) MarshalBinary() ([]byte, error)`
    MarshalBinary() returns a byte array that encodes the state of the PRNG.
    This can later be used with UnmarshalBinary() to restore the state of the
    PRNG. MarshalBinary implements the encoding.BinaryMarshaler interface.

`func (xs *Xorshift64Star) Seed(seed ...uint64)`
    Seed() implements the rand.Source interface. It provides a way to set the
    seed for the PRNG.

`func (xs *Xorshift64Star) Uint64() uint64`
    Uint64() generates the next random uint64 value.

`func (xs *Xorshift64Star) UnmarshalBinary(data []byte) error`
    UnmarshalBinary() restores the state of the PRNG from a byte array
    that was created with MarshalBinary(). UnmarshalBinary implements the
    encoding.BinaryUnmarshaler interface.

