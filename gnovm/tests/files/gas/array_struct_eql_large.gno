// MAXALLOC: 500000000
package main

// Compare two large arrays ([1000]int).
// This test should consume significantly more gas than
// array_struct_eql.gno (which uses [1]int) due to
// per-element gas metering on equality comparisons.

func main() {
	a1 := [1000]int{}
	a2 := [1000]int{}
	_ = a1 == a2
}

// Gas:
// 240243
