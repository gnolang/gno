package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint32 = 1<<32 - 1

func main() {
	println(c * 2)
}

// Error:
// main/uint32_fold_mul.gno:9:10-15: constant 8589934590 overflows Uint32Kind
