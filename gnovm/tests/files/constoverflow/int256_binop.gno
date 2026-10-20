package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int256 = 1
	println(x + 1<<255)
}

// Error:
// main/int256_binop.gno:8:10-20: bigint overflows target kind
