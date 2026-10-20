package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x uint8 = 1
	println(x + 1 << 8)
}

// Error:
// main/uint8_binop.gno:8:10-20: bigint overflows target kind
