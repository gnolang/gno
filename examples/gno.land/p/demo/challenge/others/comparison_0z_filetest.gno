package main

// both not const
func main() {
	a := 0
	b := "hi"
	println(a & b)
}

// Error:
// main/gno.land/p/demo/challenge/comparison_0z_filetest.gno:7: cannot use int as string
