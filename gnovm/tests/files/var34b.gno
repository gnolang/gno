package main

func f() {}

func main() {
	var a int
	a = f()
}

// Error:
// main/files/var34b.gno:7:2: f<VPBlock(3,0)> (no value) used as value
