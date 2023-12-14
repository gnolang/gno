package main

func main() {
	a := uint(1)
	println(uint64(1) == a)
}

// Error:
// main/files/types/0a1a0_filetest.gno:5: cannot use uint64 as uint
