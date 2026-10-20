package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint64 = 1 << 64

func main() {
	println(c)
}

// Error:
// main/uint64_const.gno:6:7-25: bigint overflows target kind
