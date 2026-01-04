package main

func main() {
	y := 1
	y <<= 1.25
}

// Error:
// main/shift_e8.gno:5:2-12: invalid operation: invalid shift count: (const (1.25 <untyped> bigdec))

// TypeCheckError:
// main/shift_e8.gno:5:8: 1.25 (untyped float constant) truncated to uint
