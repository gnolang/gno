package main

type A struct {
	X B
}

type B struct {
	X C
}

type C struct {
	X *A
}

func main() {
	var p, q A
	println(p == q)
}

// Output:
// true
