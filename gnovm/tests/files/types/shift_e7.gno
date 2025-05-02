package main

func main() {
	y := 1
	y <<= -1
}

// Error:
// main/files/types/shift_e7.gno:5:2: invalid operation: negative shift count: (-1 <untyped> bigint)

// TypeCheckError:
// main/files/types/shift_e7.gno:5:8: invalid operation: negative shift count -1 (untyped int constant)
