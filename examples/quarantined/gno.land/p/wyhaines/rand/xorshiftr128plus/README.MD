# package xorshiftr128plus // import "gno.land/p/demo/math/rand/xorshiftr128plus"

Xorshiftr128+ is a very fast psuedo-random number generation algorithm with
strong statistical properties.

The default random number algorithm in gno was ported from Go's v2 rand
implementatoon, which defaults to the PCG algorithm. This algorithm is
commonly used in language PRNG implementations because it has modest seeding
requirements, and generates statistically strong randomness.

This package provides an implementation of the Xorshiftr128+ PRNG algorithm.
This algorithm provides strong statistical performance with most seeds (just
don't seed it with zeros), and the performance of this implementation in Gno is
more than four times faster than the default PCG implementation in `math/rand`.

```
Benchmark
---------
PCG:           1000000 Uint64 generated in 15.48s
Xorshiftr128+: 1000000 Uint64 generated in 3.22s
Ratio:         x4.81 times faster than PCG
```

Use it directly:

```
prng = xorshiftr128plus.New() // pass a uint64 to seed it or pass nothing to seed it with entropy
```

Or use it as a drop-in replacement for the default PRNT in Rand:

```
source = xorshiftr128plus.New()
prng := rand.New(source)
```

## TYPES

```
type Xorshiftr128Plus struct {
	// Has unexported fields.
}
```

`func New(seeds ...uint64) *Xorshiftr128Plus`

`func (xs *Xorshiftr128Plus) MarshalBinary() ([]byte, error)`
    MarshalBinary() returns a byte array that encodes the state of the PRNG.
    This can later be used with UnmarshalBinary() to restore the state of the
    PRNG. MarshalBinary implements the encoding.BinaryMarshaler interface.

`func (x *Xorshiftr128Plus) Seed(s1, s2 uint64)`

`func (x *Xorshiftr128Plus) Uint64() uint64`

`func (xs *Xorshiftr128Plus) UnmarshalBinary(data []byte) error`
    UnmarshalBinary() restores the state of the PRNG from a byte array
    that was created with MarshalBinary(). UnmarshalBinary implements the
    encoding.BinaryUnmarshaler interface.

