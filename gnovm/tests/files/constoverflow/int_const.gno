package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int = 1 << 63

func main() {
	println(c)
}

// Error:
// main/int_const.gno:6:7-22: bigint overflows target kind
