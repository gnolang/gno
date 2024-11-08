package main

func foo() (int, bool) {
	return 1, true
}

func main() {
	var a, b int = foo()
	println(a, b)
}

// Error:
// main/files/assign33.gno:8:6: cannot use bool as int
