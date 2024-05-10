package main

type S struct {
	S
}

func main() {
	var a, b S
	println(a == b)
}

// Error:
// Cyclic dependency detected: S(File: files/recursive5.gno, Line: 3) -> S(File: files/recursive5.gno, Line: 3)
