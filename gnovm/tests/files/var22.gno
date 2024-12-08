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
