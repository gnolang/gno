package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x uint8 = -1
	println(x)
}

// Error:
// main/uint8_var_under.gno:7:6-18: bigint underflows target kind
