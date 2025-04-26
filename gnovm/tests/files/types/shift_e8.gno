package main

func main() {
	y := 1
	y <<= 1.25
}

// Error:
// main/files/types/shift_e8.gno:5:2: invalid operation: invalid shift count: (const (1.25 <untyped> bigdec))

// TypeCheckError:
// main/files/types/shift_e8.gno:5:8: 1.25 (untyped float constant) truncated to uint
