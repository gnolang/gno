package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int = 1
	println(x + 1 << 63)
}

// Error:
// main/int_binop.gno:8:10-21: bigint overflows target kind
