package main

func main() {
	println(1 << 1.25)
}

// Error:
// main/files/types/shift_e2.gno:4:10: invalid operation: invalid shift count: (const (1.25 <untyped> bigdec))

// TypeCheckError:
// main/files/types/shift_e2.gno:4:15: 1.25 (untyped float constant) truncated to uint
