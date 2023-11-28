package main

// one untyped const, one typed const
func main() {
	println("b" - "a")
}

// Error:
// main/files/types/2b3_filetest.gno:5: - should have operand number, while have: <untyped> string
