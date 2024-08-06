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
// main/files/recursive2.gno:1:1: invalid recursive type: A -> B -> C -> A
