package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint8 = 1

func main() {
	println(-c)
}

// Error:
// main/uint8_neg.gno:9:10-12: constant -1 overflows Uint8Kind
