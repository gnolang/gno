package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint8 = 1<<8 - 1

func main() {
	println(c + 1)
}

// Error:
// main/uint8_fold.gno:9:10-15: constant 256 overflows Uint8Kind
