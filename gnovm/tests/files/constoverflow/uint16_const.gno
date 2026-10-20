package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint16 = 1 << 16

func main() {
	println(c)
}

// Error:
// main/uint16_const.gno:6:7-25: bigint overflows target kind
