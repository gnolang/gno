package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func f() uint256 {
	return 1 << 256
}

func main() {
	println(f())
}

// Error:
// main/uint256_return.gno:7:2-17: bigint overflows target kind
