package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int256
	x = 1 << 255
	println(x)
}

// Error:
// main/int256_assign.gno:8:2-14: bigint overflows target kind
