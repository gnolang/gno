package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int32 = 1 << 31
	println(x)
}

// Error:
// main/int32_var.gno:7:6-23: bigint overflows target kind
