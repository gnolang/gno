package main

type Error int8

// one untyped const, one typed const
func main() {
	r := 1
	r += Error(1)
	println(r)
}

// Error:
// main/files/types/add_assign_b0.gno:8:2: invalid operation: mismatched types int and main.Error
