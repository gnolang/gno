package main

import "bytes"

// both not const, and both interface
func main() {
	// lv.T: *<databyte> uint8, rv.T: *<databyte> uint8
	cmp := bytes.Compare([]byte("a"), []byte("b"))
	println(cmp)
}

// Output:
// -1
