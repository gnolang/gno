package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int8 = -1 << 7

func main() {
	println(-c)
}

// Error:
// main/int8_neg.gno:9:10-12: constant 128 overflows Int8Kind
