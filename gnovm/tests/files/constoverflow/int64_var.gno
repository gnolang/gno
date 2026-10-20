package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int64 = 1 << 63
	println(x)
}

// Error:
// main/int64_var.gno:7:6-23: bigint overflows target kind
