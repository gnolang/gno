package main

type S struct {
	S
}

func main() {
	var a, b S
	println(a == b)
}

// Error:
// Cyclic dependency detected: S(File: files/recursive5.gno) -> S(File: files/recursive5.gno)
