package main

func main() {
	a := uint(1)
	println(a == uint64(1))
}

// Error:
// main/files/types/EQL_0a1a1.gno:5: cannot use uint64 as uint
