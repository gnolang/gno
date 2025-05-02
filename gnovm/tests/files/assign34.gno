package main

func foo() (int, bool) {
	return 1, true
}

func main() {
	var a, b = foo()
	println(a, b)
}

// Output:
// 1 true
