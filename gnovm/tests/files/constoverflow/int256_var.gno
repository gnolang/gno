package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int256 = 1 << 255
	println(x)
}

// Error:
// main/int256_var.gno:7:6-25: bigint overflows target kind
