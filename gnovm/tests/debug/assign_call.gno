package main

func foo() (int, float32) {
	return 1, 1.0
}

func main() {
	var s, s1 string
	s, s1 = foo()
	println(s)
}

// Error:
// main/debug/assign_call.gno:9: cannot use int as string
