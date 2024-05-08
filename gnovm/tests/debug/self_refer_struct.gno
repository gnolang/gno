package main

type S struct {
	T S
}

func main() {
	var a, b S
	println(a == b)
}

// Error:
// debug/self_refer_struct.gno: Cyclic dependency detected: S (Line: 3) -> S (Line: 3)
