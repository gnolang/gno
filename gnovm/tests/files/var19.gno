package main

func main() {
	a, b, c := 1, a+1
	println(a)
	println(b)
	println(c)
}

// Error:
// main/var19.gno:4:6-24: missing init expr for c<!VPInvalid(0)>

// TypeCheckError:
// main/var19.gno:4:12: missing init expr for c; main/var19.gno:4:19: undefined: a
