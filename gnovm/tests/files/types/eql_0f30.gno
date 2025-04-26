package main

// both not const, and both interface
func main() {
	println([]byte("a") == []byte("b"))
}

// Error:
// main/files/types/eql_0f30.gno:5:10: []uint8 can only be compared to nil

// TypeCheckError:
// main/files/types/eql_0f30.gno:5:10: invalid operation: []byte("a") == []byte("b") (slice can only be compared to nil)
