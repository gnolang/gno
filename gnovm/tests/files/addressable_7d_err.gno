package main

func main() {
	x := 1
	_ = &(-x)
}

// Error:
// main/addressable_7d_err.gno:5:6-11: invalid operation: cannot take address of -x<VPBlock(1,0)> (value of type int)

// TypeCheckError:
// main/addressable_7d_err.gno:5:7: invalid operation: cannot take address of (-x) (value of type int)
