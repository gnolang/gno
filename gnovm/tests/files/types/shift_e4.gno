package main

func main() {
	x := 1.25
	println(1 << x)
}

// Error:
// main/files/types/shift_e4.gno:5:10: invalid operation: invalid shift count: x<VPBlock(1,0)>

// TypeCheckError:
// main/files/types/shift_e4.gno:5:15: invalid operation: shift count x (variable of type float64) must be integer
