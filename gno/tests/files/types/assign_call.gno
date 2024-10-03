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
// main/files/types/assign_call.gno:9:2: cannot use int as string
