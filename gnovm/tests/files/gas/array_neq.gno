// MAXALLOC: 500000000
package main

// Checks inequality of two [100]int arrays.
//
// large_array_eql.gno: same test with == operator
func main() {
	a1 := [100]int{}
	a2 := [100]int{}
	_ = a1 != a2
}

// Gas:
// 17336
