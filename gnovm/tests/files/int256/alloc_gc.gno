// MAXALLOC: 2000000
package main

import "runtime"

// Each non-zero int256/uint256 result is a 64-byte allocation (a boxed
// Bits256Value); zero is stored as nil and allocates nothing. The GC recount
// must agree with what was charged, so that repeated collections under a
// small limit never report an overrun.
func main() {
	var keep [100]uint256
	var acc uint256 = 1
	for i := 0; i < 100000; i++ {
		acc = acc*3 + 1
		keep[i%100] = acc
	}
	zeros := make([]int256, 1000)
	runtime.GC()
	println(keep[0] != 0, len(zeros))

	// One value copied into many slots: each copy is charged as a box of
	// its own, as the recount counts it.
	copies := make([]uint256, 5000)
	for i := range copies {
		copies[i] = acc
	}
	runtime.GC()
	println(copies[4999] == acc)
}

// Output:
// true 1000
// true
