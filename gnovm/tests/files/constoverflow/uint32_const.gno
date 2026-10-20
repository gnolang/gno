package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c uint32 = 1 << 32

func main() {
	println(c)
}

// Error:
// main/uint32_const.gno:6:7-25: bigint overflows target kind
