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
// main/debug/self_refer_struct2.gno:1: Cyclic dependency detected: [C A B C]
