package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func f() uint8 {
	return 1 << 8
}

func main() {
	println(f())
}

// Error:
// main/uint8_return.gno:7:2-15: bigint overflows target kind
