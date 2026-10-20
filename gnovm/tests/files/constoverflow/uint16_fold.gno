package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint16 = 1<<16 - 1

func main() {
	println(c + 1)
}

// Error:
// main/uint16_fold.gno:9:10-15: constant 65536 overflows Uint16Kind
