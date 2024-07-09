package main

func foo() (int, bool) {
	return 1, true
}

func main() {
	var a, b, c = 2, foo()

	println(a, b, c)
}

// Error:
// main/files/var22.gno:8:6: assignment mismatch: 3 variable(s) but 2 value(s)
