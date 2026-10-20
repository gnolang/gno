package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int16 = -1<<15 - 1
	println(x)
}

// Error:
// main/int16_var_under.gno:7:6-26: bigint underflows target kind
