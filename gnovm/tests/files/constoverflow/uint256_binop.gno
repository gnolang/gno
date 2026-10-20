package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x uint256 = 1
	println(x + 1 << 256)
}

// Error:
// main/uint256_binop.gno:8:10-22: bigint overflows target kind
