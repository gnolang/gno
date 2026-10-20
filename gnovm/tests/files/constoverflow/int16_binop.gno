package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int16 = 1
	println(x + 1 << 15)
}

// Error:
// main/int16_binop.gno:8:10-21: bigint overflows target kind
