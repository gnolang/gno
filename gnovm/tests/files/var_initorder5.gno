package main

// Tests a single multi-name var spec where both names share the same
// initializer dependencies. A and B are declared by ONE ValueDecl so they
// are treated as a single initialization step; within the step values are
// evaluated left-to-right.
// Init order: counter (no deps), then {A,B} spec (deps counter via incr),
// then C (deps A and B).

var counter int

func incr() int { counter++; return counter }

var (
	A, B = incr(), incr()
	C    = A + B
)

func main() { println(A, B, C) }

// Output:
// 1 2 3
