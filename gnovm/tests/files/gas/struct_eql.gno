// MAXALLOC: 500000000
package main

// Test struct equality gas metering for both homogeneous
// structs and structs containing string fields.

type Point struct {
	X int
	Y int
	Z int
}

type Named struct {
	Name string
	Age  int
}

func main() {
	// Homogeneous struct (all int fields).
	p1 := Point{1, 2, 3}
	p2 := Point{1, 2, 3}
	_ = p1 == p2

	// Struct with a string field; exercises both per-field
	// and per-byte gas charging.
	buf := make([]byte, 1000)
	s := string(buf)
	n1 := Named{s, 42}
	n2 := Named{s, 42}
	_ = n1 == n2
}

// Gas:
// 175092
