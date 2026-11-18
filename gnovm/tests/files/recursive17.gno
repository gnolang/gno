package main

// A type cycle that passes through a constant: E's length needs N, N
// converts to T, and T holds E inline.

type E [N]int

const N = T(3)

type T [2]E

func main() {}

// Error:
// main/recursive17.gno:6:6-14: invalid recursive type: E -> N -> T -> E

// TypeCheckError:
// main/recursive17.gno:6:6: invalid recursive type E; main/recursive17.gno:6:6: 	E refers to N; main/recursive17.gno:8:7: 	N refers to T; main/recursive17.gno:10:6: 	T refers to E; main/recursive17.gno:8:13: cannot convert 3 (untyped int constant) to type T
