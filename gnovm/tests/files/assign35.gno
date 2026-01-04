package main

func foo() (int, bool) {
	return 1, true
}

func main() {
	a, b := 2, foo()

	println(a, b)
}

// Error:
// main/assign35.gno:8:2-18: multiple-value foo<VPBlock(3,0)> (value of type [int bool]) in single-value context

// TypeCheckError:
// main/assign35.gno:8:13: multiple-value foo() (value of type (int, bool)) in single-value context
