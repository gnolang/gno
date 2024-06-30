package main

type A struct {
	X B
}

type B struct {
	X C
}

type C struct {
	X A
}

func main() {
	var p, q A
	println(p == q)
}

// Error:
// Cyclic dependency detected: C(File: files/recursive2.gno) -> A(File: files/recursive2.gno) -> B(File: files/recursive2.gno) -> C(File: files/recursive2.gno)
