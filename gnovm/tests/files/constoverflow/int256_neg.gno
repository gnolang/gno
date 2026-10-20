package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int256 = -1 << 255

func main() {
	println(-c)
}

// Error:
// main/int256_neg.gno:9:10-12: constant 57896044618658097711785492504343953926634992332820282019728792003956564819968 overflows Int256Kind
