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
