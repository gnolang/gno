package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int16 = 1<<15 - 1

func main() {
	println(c * 2)
}

// Error:
// main/int16_fold_mul.gno:9:10-15: constant 65534 overflows Int16Kind
