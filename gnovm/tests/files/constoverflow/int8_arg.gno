package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func f(x int8) { println(x) }

func main() {
	f(1 << 7)
}

// Error:
// main/int8_arg.gno:9:2-11: bigint overflows target kind
