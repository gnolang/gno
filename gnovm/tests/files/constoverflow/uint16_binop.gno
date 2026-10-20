package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x uint16 = 1
	println(x + 1 << 16)
}

// Error:
// main/uint16_binop.gno:8:10-21: bigint overflows target kind
