package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func main() {
	var x int16 = 1
	x += 1 << 15
	println(x)
}

// Error:
// main/int16_compound.gno:8:2-14: bigint overflows target kind
