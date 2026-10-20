package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int16 = 1 << 15

func main() {
	println(c)
}

// Error:
// main/int16_const.gno:6:7-24: bigint overflows target kind
