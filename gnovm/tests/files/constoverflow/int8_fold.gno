package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int8 = 1<<7 - 1

func main() {
	println(c + 1)
}

// Error:
// main/int8_fold.gno:9:10-15: constant 128 overflows Int8Kind
