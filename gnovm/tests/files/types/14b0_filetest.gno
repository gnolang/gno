package main

type Error int8

// one untyped const, one typed const
func main() {
	r := 1
	r -= Error(1)
	println(r)
}

// Error:
// main/files/types/14b0_filetest.gno:8: invalid operation: mismatched types int and main.Error
