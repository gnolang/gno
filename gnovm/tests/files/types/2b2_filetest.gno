package main

// one untyped const, one typed const
func main() {
	println(1 - "a")
}

// Error:
// main/files/types/2b2_filetest.gno:5: - should have operand number, while have: <untyped> string
