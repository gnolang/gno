package main

type Error int8

// one untyped const, one typed const
func main() {
	r := 1
	r += Error(1)
	println(r)
}

// Error:
// main/files/types/13b0_filetest.gno:8: cannot use main.Error as int without explicit conversion
