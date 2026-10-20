package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint256 = 1<<256 - 1

func main() {
	println(c + 1)
}

// Error:
// main/uint256_fold.gno:9:10-15: constant 115792089237316195423570985008687907853269984665640564039457584007913129639936 overflows Uint256Kind
