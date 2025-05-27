package main

func r() int {
	return 1
}

func main() {
	var a, b, c = r()
}

// Error:
// main/var20.gno:8:6-19: assignment mismatch: 3 variable(s) but r<VPBlock(3,0)>() returns 1 value(s)

// TypeCheckError:
// main/var20.gno:8:16: assignment mismatch: 3 variables but r returns 1 value; main/var20.gno:8:6: declared and not used: a; main/var20.gno:8:9: declared and not used: b; main/var20.gno:8:12: declared and not used: c
