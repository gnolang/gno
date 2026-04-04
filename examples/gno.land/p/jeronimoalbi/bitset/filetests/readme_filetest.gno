package main

import "gno.land/p/jeronimoalbi/bitset"

func main() {
	var b bitset.BitSet
	b.Set(0)
	b.Set(2)
	b.Set(5)

	println("Test bit 0:", b.Test(0))
	println("Test bit 1:", b.Test(1))
	println("Test bit 2:", b.Test(2))
	println("Test bit 5:", b.Test(5))
	println("Len:", b.Len())
	println("BitSet:", b.String())
}

// Output:
// Test bit 0: true
// Test bit 1: false
// Test bit 2: true
// Test bit 5: true
// Len: 3
// BitSet: 100101
