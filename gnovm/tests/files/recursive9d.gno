package main

type S = struct {
	*S
}

func main() {}

// Error:
// main/files/recursive9d.gno:3:6: invalid recursive type: S -> S

// TypeCheckError:
// main/files/recursive9d.gno:3:6: invalid recursive type: S refers to itself
