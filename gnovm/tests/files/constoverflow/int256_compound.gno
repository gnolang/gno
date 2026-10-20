package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int256 = 1
	x += 1 << 255
	println(x)
}

// Error:
// main/int256_compound.gno:8:2-15: bigint overflows target kind
