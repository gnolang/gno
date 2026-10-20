package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int16 = 1

func main() {
	println(c << 16)
}

// Error:
// main/int16_shift.gno:8:1-10:2: constant overflows
