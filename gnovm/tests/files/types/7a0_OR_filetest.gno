package main

// both typed(different) const
func main() {
	println(int(0) | int8(1))
}

// Error:
// main/files/types/7a0_OR_filetest.gno:5: invalid operation: mismatched types int8 and int
