package main

func main() {
	a := 6
	b := "3"
	a &= b // Error: mismatched types int and string
	println(a)
}

// Error:
// main/gno.land/p/demo/challenge/comparison_3l_filetest.gno:6: mismatch type for Assign, in case: a, b = x, y
