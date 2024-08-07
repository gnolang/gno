package main

func main() {
	println(1 << float32(1.25))
}

// Error:
// main/files/types/shift_e1.gno:4:10: invalid operation: invalid shift count: (const (1.25 float32))
