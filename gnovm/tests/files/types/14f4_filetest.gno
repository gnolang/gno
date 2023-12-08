package main

// both not const, and both interface
func main() {
	"a" -= "b"
}

// Error:
// main/files/types/14f4_filetest.gno:5: operator -= not defined on: <untyped> string
