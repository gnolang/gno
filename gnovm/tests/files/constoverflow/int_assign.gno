package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int
	x = 1 << 63
	println(x)
}

// Error:
// main/int_assign.gno:8:2-13: bigint overflows target kind
