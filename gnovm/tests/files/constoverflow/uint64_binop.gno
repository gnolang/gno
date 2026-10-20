package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x uint64 = 1
	println(x + 1<<64)
}

// Error:
// main/uint64_binop.gno:8:10-19: bigint overflows target kind
