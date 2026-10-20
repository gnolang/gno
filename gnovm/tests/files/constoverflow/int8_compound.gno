package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int8 = 1
	x += 1 << 7
	println(x)
}

// Error:
// main/int8_compound.gno:8:2-13: bigint overflows target kind
