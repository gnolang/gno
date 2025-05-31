package main

func foo() (int, bool) {
	return 1, true
}

func main() {
	var a, b, c = 2, foo()

	println(a, b, c)
}

// Error:
// main/var22.gno:8:6-24: missing init expr for c<!VPInvalid(0)>

// TypeCheckError:
// main/var22.gno:8:12: missing init expr for c; main/var22.gno:8:19: multiple-value foo() (value of type (int, bool)) in single-value context
