package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func f(x uint8) { println(x) }

func main() {
	f(1 << 8)
}

// Error:
// main/uint8_arg.gno:9:2-11: bigint overflows target kind
