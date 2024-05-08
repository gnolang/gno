package main

type S struct {
	T S
}

func main() {
	var a, b S
	println(a == b)
}

// Error:
// main/debug/self_refer_struct.gno:1: Cyclic dependency detected: [S S]
