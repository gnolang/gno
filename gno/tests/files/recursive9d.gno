package main

type S = struct {
	*S
}

func main() {}

// Error:
// main/files/recursive9d.gno:1:1: invalid recursive type: S -> S
