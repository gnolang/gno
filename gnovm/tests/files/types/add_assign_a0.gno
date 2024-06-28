package main

// both typed(different) const
func main() {
	int(0) += int8(1)
}

// Error:
// main/files/types/add_assign_a0.gno:5: invalid operation: mismatched types int and int8
