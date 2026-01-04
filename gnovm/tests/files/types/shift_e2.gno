package main

func main() {
	println(1 << 1.25)
}

// Error:
// main/shift_e2.gno:4:10-19: invalid operation: invalid shift count: (const (1.25 <untyped> bigdec))

// TypeCheckError:
// main/shift_e2.gno:4:15: 1.25 (untyped float constant) truncated to uint
