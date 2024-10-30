package main

func foo()  {}

func main() {
	x := foo()
}

// Error:
// main/files/assign32.gno:6:2: foo<VPBlock(3,0)>() (no value) used as value
