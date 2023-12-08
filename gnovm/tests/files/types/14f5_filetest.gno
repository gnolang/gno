package main

// both not const, and both interface
func main() {
	var a string = "a"
	a -= "b"
}

// Error:
// main/files/types/14f5_filetest.gno:6: operator -= not defined on: string
