// MAXALLOC: 500000000
package main

// Test nested array equality ([N][M]T) to verify that
// per-element gas metering applies recursively.

func main() {
	a1 := [10][10]int{}
	a2 := [10][10]int{}
	_ = a1 == a2
}

// Gas:
// 36635
