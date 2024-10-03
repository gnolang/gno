package main

func foo() (int, bool) {
	return 1, true
}

func main() {
	var a, b, c = 2, 3, 4, foo()

	println(a, b, c)
}

// Error:
// main/files/var22b.gno:8:6: extra init expr foo<VPUverse(0)>()
