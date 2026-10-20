package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int256 = -1<<255 - 1
	println(x)
}

// Error:
// main/int256_var_under.gno:7:6-28: bigint underflows target kind
