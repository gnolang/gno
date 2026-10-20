package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func f() uint32 {
	return 1 << 32
}

func main() {
	println(f())
}

// Error:
// main/uint32_return.gno:7:2-16: bigint overflows target kind
