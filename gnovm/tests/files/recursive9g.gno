package main

// The function type breaks the cycle's infinite size, but every type in the
// cycle is an alias, so the chain never resolves to a concrete type.

type A = func() A

func main() {}

// Error:
// main/recursive9g.gno:6:6-18: invalid recursive type: A -> A

// TypeCheckError:
// main/recursive9g.gno:6:6: invalid recursive type: A refers to itself
