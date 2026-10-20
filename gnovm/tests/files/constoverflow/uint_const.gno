package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint = 1 << 64

func main() {
	println(c)
}

// Error:
// main/uint_const.gno:6:7-23: bigint overflows target kind
