package main

func main() {
	a, b := 1, 2
	_ = &(a + b)
}

// Error:
// main/addressable_7c_err.gno:5:6-14: invalid operation: cannot take address of a<VPBlock(1,0)> + b<VPBlock(1,1)> (value of type int)

// TypeCheckError:
// main/addressable_7c_err.gno:5:7: invalid operation: cannot take address of (a + b) (value of type int)
