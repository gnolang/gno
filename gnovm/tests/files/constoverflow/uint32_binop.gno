package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x uint32 = 1
	println(x + 1 << 32)
}

// Error:
// main/uint32_binop.gno:8:10-21: bigint overflows target kind
