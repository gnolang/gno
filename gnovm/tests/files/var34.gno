package main

func f() {}

func main() {
	var t = f()
}

// Error:
// main/files/var34.gno:6:6: f<VPBlock(3,0)> (no value) used as value
