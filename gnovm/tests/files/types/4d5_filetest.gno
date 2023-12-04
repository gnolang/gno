package main

// both untyped const
// TODO: dec value representation
func main() {
	a := 1 / 0
	println(a)
}

// Error:
// main/files/types/4d5_filetest.gno:6: invalid operation: division by zero
