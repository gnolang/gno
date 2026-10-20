package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int32
	x = 1 << 31
	println(x)
}

// Error:
// main/int32_assign.gno:8:2-13: bigint overflows target kind
