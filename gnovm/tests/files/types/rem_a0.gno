package main

// both typed(different) const
func main() {
	println(int(1) % int8(1))
}

// Error:
// main/files/types/rem_a0.gno:5:10: invalid operation: mismatched types int and int8

// TypeCheckError:
// main/files/types/rem_a0.gno:5:10: invalid operation: int(1) % int8(1) (mismatched types int and int8)
