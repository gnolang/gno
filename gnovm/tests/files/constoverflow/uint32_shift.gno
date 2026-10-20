package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint32 = 1

func main() {
	println(c << 32)
}

// Error:
// main/uint32_shift.gno:8:1-10:2: constant overflows
