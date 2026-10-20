package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint16 = 1

func main() {
	println(c << 16)
}

// Error:
// main/uint16_shift.gno:8:1-10:2: constant overflows
