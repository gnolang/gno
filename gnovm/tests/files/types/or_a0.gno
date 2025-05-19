package main

// both typed(different) const
func main() {
	println(int(0) | int8(1))
}

// Error:
// main/or_a0.gno:5:10-26: invalid operation: (const (0 int)) | (const (1 int8)) (mismatched types int and int8)

// TypeCheckError:
// main/or_a0.gno:5:10: invalid operation: int(0) | int8(1) (mismatched types int and int8)
