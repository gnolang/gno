package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint = 1<<64 - 1

func main() {
	println(c * 2)
}

// Error:
// main/uint_fold_mul.gno:9:10-15: constant 36893488147419103230 overflows UintKind
