package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint256 = 1 << 256

func main() {
	println(c)
}

// Error:
// main/uint256_const.gno:6:7-27: bigint overflows target kind
