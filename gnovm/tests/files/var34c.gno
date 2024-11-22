package main

func f() {}

func main() {
	var a, b int = f(), 1
}

// Error:
// main/files/var34b.gno:7:2: f<VPBlock(3,0)> (no value) used as value
