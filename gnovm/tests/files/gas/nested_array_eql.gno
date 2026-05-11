// MAXALLOC: 500000000
package main

// Checks equality of two [10][10]int nested arrays (100 leaf elements).
//
// large_array_eql.gno: same 100 elements as flat [100]int
func main() {
	a1 := [10][10]int{}
	a2 := [10][10]int{}
	_ = a1 == a2
}

// Gas:
// 22162
