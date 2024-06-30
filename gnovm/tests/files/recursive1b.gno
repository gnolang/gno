package main

type S struct {
	T S
}

func main() {
	var a, b S
}

// Error:
// Cyclic dependency detected: S(File: files/recursive1b.gno, Line: 3) -> S(File: files/recursive1b.gno, Line: 3)
