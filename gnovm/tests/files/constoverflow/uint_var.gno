package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x uint = 1 << 64
	println(x)
}

// Error:
// main/uint_var.gno:7:6-22: bigint overflows target kind
