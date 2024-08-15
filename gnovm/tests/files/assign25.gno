package main

func foo() (int, bool) {
	return 1, true
}

func main() {
	a, b, c := 2, foo()

	println(a, b, c)
}

// Error:
// main/files/assign25.gno:8:2: assignment mismatch: 3 variable(s) but 2 value(s)
