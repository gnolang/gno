# MurmurHash3 Package

Package implements Austin Appleby's MurmurHash3 algorithm.

MurmurHash3 is a fast, non-cryptographic hash function well suited for hash tables,
bloom filters, data partitioning, and deduplication where speed and good distribution
matter more than cryptographic security.

## API

- `New32()` — Returns a `hash.Hash32` with seed zero
- `NewWithSeed32(seed)` — Returns a `hash.Hash32` with the given seed
- `Sum32(data)` — One-shot hash with seed zero
- `Sum32WithSeed(data, seed)` — One-shot hash with the given seed
- `New64()` — Returns a `hash.Hash64` using seeds zero and one
- `NewWithSeed64(seed1, seed2)` — Returns a `hash.Hash64` with the given seeds
- `Sum64(data)` — One-shot 64-bit hash using seeds zero and one
- `Sum64WithSeed(data, seed1, seed2)` — One-shot 64-bit hash with the given seeds
- `EncodeToString(sum64)` — Returns the hexadecimal encoding of a hash value

## Usage

[embedmd]:# (filetests/readme_filetest.gno go)
```go
package main

import "gno.land/p/jeronimoalbi/murmur3"

func main() {
	data := []byte("Hello, world!")
	seed := uint32(42)

	// Hash32 without seed
	h32 := murmur3.New32()
	h32.Write(data)
	sum32 := uint64(h32.Sum32())
	println(murmur3.EncodeToString(sum32))

	// Hash32 with seed
	h32 = murmur3.NewWithSeed32(seed)
	h32.Write(data)
	sum32 = uint64(h32.Sum32())
	println(murmur3.EncodeToString(sum32))

	// Hash32 without seed using a helper function
	sum32 = uint64(murmur3.Sum32(data))
	println(murmur3.EncodeToString(sum32))

	// Hash32 with seed using a helper function
	sum32 = uint64(murmur3.Sum32WithSeed(data, seed))
	println(murmur3.EncodeToString(sum32))

	// Hash64 without seed
	h64 := murmur3.New64()
	h64.Write(data)
	sum64 := h64.Sum64()
	println(murmur3.EncodeToString(sum64))

	// Hash64 with seed
	h64 = murmur3.NewWithSeed64(0, seed)
	h64.Write(data)
	sum64 = h64.Sum64()
	println(murmur3.EncodeToString(sum64))

	// Hash64 without seed using a helper function
	sum64 = murmur3.Sum64(data)
	println(murmur3.EncodeToString(sum64))

	// Hash64 with seed using a helper function
	sum64 = murmur3.Sum64WithSeed(data, 0, seed)
	println(murmur3.EncodeToString(sum64))
}

// Output:
// c0363e43
// 2c8c8533
// c0363e43
// 2c8c8533
// c0363e43aa5dc85b
// c0363e432c8c8533
// c0363e43aa5dc85b
// c0363e432c8c8533
```
