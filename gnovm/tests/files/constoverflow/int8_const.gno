package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int8 = 1 << 7

func main() {
	println(c)
}

// Error:
// main/int8_const.gno:6:7-22: bigint overflows target kind
