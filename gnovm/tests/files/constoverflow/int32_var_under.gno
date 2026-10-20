package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int32 = -1<<31 - 1
	println(x)
}

// Error:
// main/int32_var_under.gno:7:6-26: bigint underflows target kind
