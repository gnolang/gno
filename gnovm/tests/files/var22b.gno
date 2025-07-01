package main

func foo() (int, bool) {
	return 1, true
}

func main() {
	var a, b, c = 2, 3, 4, foo()

	println(a, b, c)
}

// Error:
// main/var22b.gno:8:6-30: extra init expr foo<VPInvalid(0)>()

// TypeCheckError:
// main/var22b.gno:8:25: extra init expr foo()
