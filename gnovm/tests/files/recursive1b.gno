package main

type S struct {
	T S
}

func main() {
	var a, b S
}

// Error:
// Cyclic dependency detected: S(File: files/recursive1b.gno) -> S(File: files/recursive1b.gno)
