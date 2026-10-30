package main

// The method signature breaks the cycle's infinite size, but every type in
// the cycle is an alias, so the chain never resolves to a concrete type.

type A = interface{ M() A }

func main() {}

// Error:
// main/recursive9h.gno:6:6-28: invalid recursive type: A -> A

// TypeCheckError:
// main/recursive9h.gno:6:6: invalid recursive type: A refers to itself
