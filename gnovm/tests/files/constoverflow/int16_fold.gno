package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int16 = 1<<15 - 1

func main() {
	println(c + 1)
}

// Error:
// main/int16_fold.gno:9:10-15: constant 32768 overflows Int16Kind
