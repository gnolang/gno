package main

type S struct {
	T S
}

func main() {
	var a, b S
	println(a == b)
}

// Error:
// main/recursive1.gno:3:6-5:2: invalid recursive type: S -> S

// TypeCheckError:
// main/recursive1.gno:3:6: invalid recursive type: S refers to itself
