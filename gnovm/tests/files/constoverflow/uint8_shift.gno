package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint8 = 1

func main() {
	println(c << 8)
}

// Error:
// main/uint8_shift.gno:8:1-10:2: constant overflows
