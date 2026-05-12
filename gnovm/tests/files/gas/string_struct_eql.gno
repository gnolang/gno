// MAXALLOC: 500000000
package main

// Checks equality of two structs containing a 1000-byte string field.
//
// large_struct_eql.gno: struct with only int fields
type Named struct {
	Name string
	Age  int
}

func main() {
	buf := make([]byte, 1000)
	s := string(buf)
	n1 := Named{s, 42}
	n2 := Named{s, 42}
	_ = n1 == n2
}

// Gas:
// 12845
