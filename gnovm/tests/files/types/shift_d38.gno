package main

func main() {
	println(1 << nil)
}

// Error:
// main/files/types/shift_d38.gno:3:1: cannot convert (undefined) to uint
