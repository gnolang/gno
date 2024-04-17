package main

func main() {
	a := uint(1)
	println(a == uint64(1))
}

// Error:
// main/files/types/0a1a1_filetest.gno:5: cannot use uint64 as uint
