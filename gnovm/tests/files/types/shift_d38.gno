package main

func main() {
	println(1 << nil)
}

// Error:
// main/files/types/shift_d38.gno:4:10: invalid operation: invalid shift count: (const (undefined))
