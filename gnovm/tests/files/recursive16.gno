package main

// A cycle through an array-length constant: T's length needs M, and M
// converts to T. The pre-scan does not follow values; the dependency walk
// reports it.

const M = T(0)

type T [M]int

func main() {}

// Error:
// main/recursive16.gno:9:6-14: invalid recursive value: T -> M -> T

// TypeCheckError:
// main/recursive16.gno:7:7: invalid cycle in declaration of M; main/recursive16.gno:7:7: 	M refers to T; main/recursive16.gno:9:6: 	T refers to M
