package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint8 = 1 << 8

func main() {
	println(c)
}

// Error:
// main/uint8_const.gno:6:7-23: bigint overflows target kind
