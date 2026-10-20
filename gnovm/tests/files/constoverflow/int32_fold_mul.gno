package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int32 = 1<<31 - 1

func main() {
	println(c * 2)
}

// Error:
// main/int32_fold_mul.gno:9:10-15: constant 4294967294 overflows Int32Kind
