package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x uint64 = 1 << 64
	println(x)
}

// Error:
// main/uint64_var.gno:7:6-24: bigint overflows target kind
