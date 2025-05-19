package main

type S struct {
	S
}

func main() {
	var a, b S
	println(a == b)
}

// Error:
// main/recursive5.gno:3:6-5:2: invalid recursive type: S -> S

// TypeCheckError:
// main/recursive5.gno:3:6: invalid recursive type: S refers to itself
