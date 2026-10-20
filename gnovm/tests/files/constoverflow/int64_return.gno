package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func f() int64 {
	return 1 << 63
}

func main() {
	println(f())
}

// Error:
// main/int64_return.gno:7:2-16: bigint overflows target kind
