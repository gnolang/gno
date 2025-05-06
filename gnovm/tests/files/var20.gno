package main

func r() int {
	return 1
}

func main() {
	var a, b, c = r()
}

// Error:
// main/files/var20.gno:8:6: assignment mismatch: 3 variable(s) but r<VPBlock(3,0)>() returns 1 value(s)

// TypeCheckError:
// main/files/var20.gno:8:16: assignment mismatch: 3 variables but r returns 1 value; main/files/var20.gno:8:6: declared and not used: a; main/files/var20.gno:8:9: declared and not used: b; main/files/var20.gno:8:12: declared and not used: c
