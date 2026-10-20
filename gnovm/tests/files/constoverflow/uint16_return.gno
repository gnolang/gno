package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func f() uint16 {
	return 1 << 16
}

func main() {
	println(f())
}

// Error:
// main/uint16_return.gno:7:2-16: bigint overflows target kind
