package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint64 = 1<<64 - 1

func main() {
	println(c + 1)
}

// Error:
// main/uint64_fold.gno:9:10-15: constant 18446744073709551616 overflows Uint64Kind
