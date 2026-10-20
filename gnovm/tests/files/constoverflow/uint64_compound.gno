package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x uint64 = 1
	x += 1 << 64
	println(x)
}

// Error:
// main/uint64_compound.gno:8:2-14: bigint overflows target kind
