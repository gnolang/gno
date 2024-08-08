package main

func main() {
	a := float64(-100)
	println(1 << a)
}

// Error:
// main/files/types/shift_e7a.gno:5:10: invalid operation: invalid shift count: a<VPBlock(1,0)>
