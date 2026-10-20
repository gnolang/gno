package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int64 = 1 << 63

func main() {
	println(c)
}

// Error:
// main/int64_const.gno:6:7-24: bigint overflows target kind
