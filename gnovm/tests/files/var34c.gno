package main

func f() {}

func main() {
	var a, b int = f(), 1
}

// Error:
// main/files/var34c.gno:6:6: f<VPBlock(3,0)> (no value) used as value
