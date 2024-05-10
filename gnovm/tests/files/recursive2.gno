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
// Cyclic dependency detected: C(File: files/recursive2.gno, Line: 11) -> A(File: files/recursive2.gno, Line: 3) -> B(File: files/recursive2.gno, Line: 7) -> C(File: files/recursive2.gno, Line: 11)
