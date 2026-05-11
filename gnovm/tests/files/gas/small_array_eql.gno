// MAXALLOC: 500000000
package main

// Checks equality of two [1]int arrays.
//
// large_array_eql.gno: same test with [100]int
func main() {
	a1 := [1]int{}
	a2 := [1]int{}
	_ = a1 == a2
}

// Gas:
// 4983
