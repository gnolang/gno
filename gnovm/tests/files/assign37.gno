package main

func f() { }

func main() {
	a := f()
}

// Error:
// main/files/assign37.gno:6:2: f<VPBlock(3,0)> (no value) used as value
