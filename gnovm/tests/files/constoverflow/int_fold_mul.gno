package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int = 1<<63 - 1

func main() {
	println(c * 2)
}

// Error:
// main/int_fold_mul.gno:9:10-15: constant 18446744073709551614 overflows IntKind
