package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x uint64
	x = 1 << 64
	println(x)
}

// Error:
// main/uint64_assign.gno:8:2-13: bigint overflows target kind
