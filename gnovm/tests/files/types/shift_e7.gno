package main

func main() {
	y := 1
	y <<= -1
}

// Error:
// main/shift_e7.gno:5:2-10: invalid operation: negative shift count: (-1 <untyped> bigint)

// TypeCheckError:
// main/shift_e7.gno:5:8: invalid operation: negative shift count -1 (untyped int constant)
