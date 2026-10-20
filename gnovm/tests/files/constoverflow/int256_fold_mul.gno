package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int256 = 1<<255 - 1

func main() {
	println(c * 2)
}

// Error:
// main/int256_fold_mul.gno:9:10-15: constant 115792089237316195423570985008687907853269984665640564039457584007913129639934 overflows Int256Kind
