# package isaac // import "gno.land/p/demo/math/rand/isaac"

This is a port of the ISAAC cryptographically secure PRNG,
originally based on the reference implementation found at
https://burtleburtle.net/bob/rand/isaacafa.html

ISAAC has excellent statistical properties, with long cycle times, and
uniformly distributed, unbiased, and unpredictable number generation. It can
not be distinguished from real random data, and in three decades of scrutiny,
no practical attacks have been found.

The default random number algorithm in gno was ported from Go's v2 rand
implementatoon, which defaults to the PCG algorithm. This algorithm is
commonly used in language PRNG implementations because it has modest seeding
requirements, and generates statistically strong randomness.

This package provides an implementation of the 32-bit ISAAC PRNG algorithm. This
algorithm provides very strong statistical performance, and is cryptographically
secure, while still being substantially faster than the default PCG
implementation in `math/rand`. Note that this package does implement a `Uint64()`
function in order to generate a 64 bit number out of two 32 bit numbers. Doing this
makes the generator only slightly faster than PCG, however,

Note that the approach to seeing with ISAAC is very important for best results,
and seeding with ISAAC is not as simple as seeding with a single uint64 value.
The ISAAC algorithm requires a 256-element seed. If used for cryptographic
purposes, this will likely require entropy generated off-chain for actual
cryptographically secure seeding. For other purposes, however, one can utilize
the built-in seeding mechanism, which will leverage the xorshiftr128plus PRNG to
generate any missing seeds if fewer than 256 are provided.


```
Benchmark
---------
PCG:         1000000 Uint64 generated in 15.58s
ISAAC:       1000000 Uint64 generated in 13.23s (uint64)
ISAAC:       1000000 Uint32 generated in 6.43s (uint32)
Ratio:       x1.18 times faster than PCG (uint64)
Ratio:       x2.42 times faster than PCG (uint32)
```

Use it directly:

```
prng = isaac.New() // pass 0 to 256 uint32 seeds; if fewer than 256 are provided, the rest
                   // will be generated using the xorshiftr128plus PRNG.
```

Or use it as a drop-in replacement for the default PRNT in Rand:

```
source = isaac.New()
prng := rand.New(source)
```

# TYPES

`
type ISAAC struct {
	// Has unexported fields.
}
`

`func New(seeds ...uint32) *ISAAC`
    ISAAC requires a large, 256-element seed. This implementation will leverage
    the entropy package combined with the the xorshiftr128plus PRNG to generate
    any missing seeds of fewer than the required number of arguments are
    provided.

`func (isaac *ISAAC) MarshalBinary() ([]byte, error)`
    MarshalBinary() returns a byte array that encodes the state of the PRNG.
    This can later be used with UnmarshalBinary() to restore the state of the
    PRNG. MarshalBinary implements the encoding.BinaryMarshaler interface.

`func (isaac *ISAAC) Seed(seed [256]uint32)`

`func (isaac *ISAAC) Uint32() uint32`

`func (isaac *ISAAC) Uint64() uint64`

`func (isaac *ISAAC) UnmarshalBinary(data []byte) error`
    UnmarshalBinary() restores the state of the PRNG from a byte array
    that was created with MarshalBinary(). UnmarshalBinary implements the
    encoding.BinaryUnmarshaler interface.

