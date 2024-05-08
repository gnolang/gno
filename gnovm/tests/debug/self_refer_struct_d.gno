package main

type S struct {
	S
}

func main() {
	var a, b S
	println(a == b)
}

// Error:
// main/debug/self_refer_struct_d.gno:1: Cyclic dependency detected: [S S]
