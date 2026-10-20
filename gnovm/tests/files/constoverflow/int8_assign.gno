package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int8
	x = 1 << 7
	println(x)
}

// Error:
// main/int8_assign.gno:8:2-12: bigint overflows target kind
