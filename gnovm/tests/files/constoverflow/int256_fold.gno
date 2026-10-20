package main

// go/types' overflow error is dropped (isDeferredOverflowError), so Gno's
// preprocessor must reject this constant overflow itself.

const c int256 = 1<<255 - 1

func main() {
	println(c + 1)
}

// Error:
// main/int256_fold.gno:9:10-15: constant 57896044618658097711785492504343953926634992332820282019728792003956564819968 overflows Int256Kind
