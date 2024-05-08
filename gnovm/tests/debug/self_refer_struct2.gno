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
// debug/self_refer_struct2.gno: Cyclic dependency detected: C (Line: 11) -> A (Line: 3) -> B (Line: 7) -> C (Line: 11)
