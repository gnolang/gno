package main

type S struct {
	T S
}

func main() {
	var a, b S
	println(a == b)
}

// Error:
// Cyclic dependency detected: S(File: files/recursive1.gno) -> S(File: files/recursive1.gno)
