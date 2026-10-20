package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func f() int256 {
	return 1 << 255
}

func main() {
	println(f())
}

// Error:
// main/int256_return.gno:7:2-17: bigint overflows target kind
