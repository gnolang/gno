package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x uint256
	x = 1 << 256
	println(x)
}

// Error:
// main/uint256_assign.gno:8:2-14: bigint overflows target kind
