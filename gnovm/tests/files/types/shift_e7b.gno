package main

func main() {
	println(1 << float64(-1))
}

// Error:
// main/files/types/shift_e7b.gno:4:10: invalid operation: invalid shift count: (const (-1 float64))

// TypeCheckError:
// main/files/types/shift_e7b.gno:4:15: invalid operation: negative shift count float64(-1) (constant -1 of type float64)
