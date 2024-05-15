package main

type S struct {
	S
}

func main() {
	var a, b S
	println(a == b)
}

// Error:
// Cyclic dependency detected: S(File: debug/self_refer_struct_d.gno, Line: 3) -> S(File: debug/self_refer_struct_d.gno, Line: 3)
