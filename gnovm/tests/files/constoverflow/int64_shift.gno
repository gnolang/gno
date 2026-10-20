package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int64 = 1

func main() {
	println(c << 64)
}

// Error:
// main/int64_shift.gno:8:1-10:2: constant overflows
