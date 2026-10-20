package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

func f(x uint256) { println(x) }

func main() {
	f(1 << 256)
}

// Error:
// main/uint256_arg.gno:9:2-13: bigint overflows target kind
