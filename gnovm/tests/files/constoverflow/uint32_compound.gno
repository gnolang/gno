package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x uint32 = 1
	x += 1 << 32
	println(x)
}

// Error:
// main/uint32_compound.gno:8:2-14: bigint overflows target kind
