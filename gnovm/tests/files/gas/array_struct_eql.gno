// MAXALLOC: 500000000
package main

// Baseline: compare two small arrays ([1]int).
// The gas consumed here serves as a control to compare against
// array_struct_eql_large.gno which uses [1000]int arrays.
// The difference in gas between the two tests demonstrates
// that per-element gas metering is working correctly.

func main() {
	a1 := [1]int{}
	a2 := [1]int{}
	_ = a1 == a2
}

// Gas:
// 5479
