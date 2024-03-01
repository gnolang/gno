package main

// both typed(different) const
func main() {
	println(int(1) * int8(1))
}

// Error:
// main/files/types/3a0_MUL_filetest.gno:5: invalid operation: mismatched types int and int8
