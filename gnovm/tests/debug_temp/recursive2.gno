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
// Cyclic dependency detected: C(File: debug/self_refer_struct2.gno, Line: 11) -> A(File: debug/self_refer_struct2.gno, Line: 3) -> B(File: debug/self_refer_struct2.gno, Line: 7) -> C(File: debug/self_refer_struct2.gno, Line: 11)
