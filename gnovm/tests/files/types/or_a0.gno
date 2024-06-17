package main

// both typed(different) const
func main() {
	println(int(0) | int8(1))
}

// Error:
// main/files/types/or_a0.gno:5: invalid operation: mismatched types int and int8
