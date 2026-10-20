package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x uint256 = 1 << 256
	println(x)
}

// Error:
// main/uint256_var.gno:7:6-26: bigint overflows target kind
