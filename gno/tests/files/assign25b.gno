package main

func foo() (int, bool) {
	return 1, true
}

func main() {
	a, b, c := 2, 3, 4, foo()

	println(a, b, c)
}

// Error:
// main/files/assign25b.gno:8:2: assignment mismatch: 3 variable(s) but 4 value(s)
