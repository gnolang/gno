package main

func main() {
	println(1.0 << float32(2.58485848))
}

// Error:
// main/files/types/shift_e1.gno:4:10: invalid operation: invalid shift count: (const (2.5848584 float32))
