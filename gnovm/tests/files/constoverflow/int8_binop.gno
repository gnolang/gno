package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int8 = 1
	println(x + 1<<7)
}

// Error:
// main/int8_binop.gno:8:10-18: bigint overflows target kind
