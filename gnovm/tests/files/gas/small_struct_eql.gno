// MAXALLOC: 500000000
package main

// Checks equality of two structs with 3 int fields.
//
// large_struct_eql.gno: same test with 20 fields
type Point struct {
	X int
	Y int
	Z int
}

func main() {
	p1 := Point{1, 2, 3}
	p2 := Point{1, 2, 3}
	_ = p1 == p2
}

// Gas:
// 6598
