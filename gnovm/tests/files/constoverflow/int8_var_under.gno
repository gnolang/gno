package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int8 = -1<<7 - 1
	println(x)
}

// Error:
// main/int8_var_under.gno:7:6-24: bigint underflows target kind
