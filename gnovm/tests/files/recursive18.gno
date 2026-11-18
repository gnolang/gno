package main

// A type cycle through a variable: E's length needs x's type, x's
// initializer names g, and g's signature names E. The dependency walk
// reports it; the pre-scan sees only type declarations.

type E [len(x)]int

var x [3]T = g(E{})

func g(E) [3]T { return [3]T{} }

type T int

func main() {}

// Error:
// main/recursive18.gno:7:6-19: invalid recursive type: E -> x -> g -> E

// TypeCheckError:
// main/recursive18.gno:7:6: invalid recursive type E; main/recursive18.gno:7:6: 	E refers to x; main/recursive18.gno:9:5: 	x refers to g; main/recursive18.gno:11:6: 	g refers to E; main/recursive18.gno:7:6: invalid recursive type E; main/recursive18.gno:7:6: 	E refers to x; main/recursive18.gno:9:5: 	x refers to E
