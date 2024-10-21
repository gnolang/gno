package main

func main() {
	x := 1.25
	println(1 << x)
}

// Error:
// main/files/types/shift_e3.gno:5:10: invalid operation: invalid shift count: x<VPBlock(1,0)>
