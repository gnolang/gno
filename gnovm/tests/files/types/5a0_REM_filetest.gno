package main

// both typed(different) const
func main() {
	println(int(1) % int8(1))
}

// Error:
// main/files/types/5a0_REM_filetest.gno:5: invalid operation: mismatched types int and int8
