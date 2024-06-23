package main

func r() int {
	return 1
}

func main() {
	var a, b, c = r()
}

// Error:
// main/files/var20.gno:8: assignment mismatch: 3 variable(s) but r<VPBlock(3,0)> returns 1 value(s)
