package main

func main() {
	a := uint(1)
	println(uint64(1) == a)
}

// Error:
// main/files/types/eql_0a1a0.gno:5: cannot use uint64 as uint
