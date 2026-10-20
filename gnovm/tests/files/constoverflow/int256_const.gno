package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int256 = 1 << 255

func main() {
	println(c)
}

// Error:
// main/int256_const.gno:6:7-26: bigint overflows target kind
