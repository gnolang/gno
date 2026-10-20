package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int16 = -1 << 15

func main() {
	println(-c)
}

// Error:
// main/int16_neg.gno:9:10-12: constant 32768 overflows Int16Kind
