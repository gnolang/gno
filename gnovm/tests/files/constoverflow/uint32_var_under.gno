package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x uint32 = -1
	println(x)
}

// Error:
// main/uint32_var_under.gno:7:6-19: bigint underflows target kind
