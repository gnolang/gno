package main

func pass(x int) (int, bool) {
	return x, true
}

func main() {
	a := 1
	a, ok := pass(a)
	println(a, ok)
}

// Output:
// 1 true
