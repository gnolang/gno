package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int8 = 1<<7 - 1

func main() {
	println(c * 2)
}

// Error:
// main/int8_fold_mul.gno:9:10-15: constant 254 overflows Int8Kind
