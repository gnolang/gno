package main

// both typed(different) const
func main() {
	int(0) += int8(1)
}

// Error:
// main/add_assign_a0.gno:5:2-19: invalid operation: mismatched types int and int8

// TypeCheckError:
// main/add_assign_a0.gno:5:2: invalid operation: int(0) += int8(1) (mismatched types int and int8)
