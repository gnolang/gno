package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func f(x int32) { println(x) }

func main() {
	f(1 << 31)
}

// Error:
// main/int32_arg.gno:9:2-12: bigint overflows target kind
