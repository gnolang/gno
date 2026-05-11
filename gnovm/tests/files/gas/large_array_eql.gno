// MAXALLOC: 500000000
package main

// Checks equality of two [100]int arrays.
//
// small_array_eql.gno: same test with [1]int
// array_neq.gno: same test with != operator
// nested_array_eql.gno: same 100 leaf elements as [10][10]int
func main() {
	a1 := [100]int{}
	a2 := [100]int{}
	_ = a1 == a2
}

// Gas:
// 17348
