package main

// both not const, and both interface
func main() {
	println([]byte("a") == []byte("b"))
}

// Error:
// main/files/types/0f30_filetest.gno:5: []uint8 can only be compared to nil
