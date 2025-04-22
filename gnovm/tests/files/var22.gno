package main

func foo() (int, bool) {
	return 1, true
}

func main() {
	var a, b, c = 2, foo()

	println(a, b, c)
}

// Error:
// main/files/var22.gno:8:6: missing init expr for c<!VPUverse(0)>

// TypeCheckError:
// main/files/var22.gno:8:12: missing init expr for c; main/files/var22.gno:8:19: multiple-value foo() (value of type (int, bool)) in single-value context
