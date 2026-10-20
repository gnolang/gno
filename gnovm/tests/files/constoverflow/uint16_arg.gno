package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func f(x uint16) { println(x) }

func main() {
	f(1 << 16)
}

// Error:
// main/uint16_arg.gno:9:2-12: bigint overflows target kind
