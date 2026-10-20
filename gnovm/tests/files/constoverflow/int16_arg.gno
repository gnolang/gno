package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func f(x int16) { println(x) }

func main() {
	f(1 << 15)
}

// Error:
// main/int16_arg.gno:9:2-12: bigint overflows target kind
