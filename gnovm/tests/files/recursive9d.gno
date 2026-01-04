package main

type S = struct {
	*S
}

func main() {}

// Error:
// main/recursive9d.gno:3:6-5:2: invalid recursive type: S -> S

// TypeCheckError:
// main/recursive9d.gno:3:6: invalid recursive type: S refers to itself
